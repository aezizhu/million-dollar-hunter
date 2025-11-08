/**
 * API client for Million Dollar Hunter frontend.
 * Copyright (c) 2025 aezizhu. All rights reserved.
 * Author: aezizhu
 * Repository: github.com/aezizhu/million-dollar-hunter
 */
import axios, { AxiosError } from 'axios';
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterResponse,
  TokenRefreshResponse,
  PortfolioListResponse,
  AddWalletRequest,
  AddWalletResponse,
  WalletView,
  TransactionPage,
  TopHoldersResponse,
  ApiError,
  Chain,
} from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_BASE_URL || 'http://localhost:8080';

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('accessToken');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiError>) => {
    const originalRequest = error.config;
    
    if (error.response?.status === 401 && originalRequest && !originalRequest.headers['X-Retry']) {
      originalRequest.headers['X-Retry'] = 'true';
      
      try {
        if (typeof window !== 'undefined') {
          const refreshToken = localStorage.getItem('refreshToken');
          const { data } = await axios.post<TokenRefreshResponse>(
            `${API_BASE_URL}/api/v1/auth/refresh`,
            {},
            { 
              headers: { 
                Authorization: refreshToken ? `Bearer ${refreshToken}` : undefined 
              } 
            }
          );
          
          localStorage.setItem('accessToken', data.accessToken);
          if (data.refreshToken) {
            localStorage.setItem('refreshToken', data.refreshToken);
          }
          originalRequest.headers.Authorization = `Bearer ${data.accessToken}`;
          
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        if (typeof window !== 'undefined') {
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          window.location.href = '/auth/login';
        }
        return Promise.reject(refreshError);
      }
    }
    
    return Promise.reject(error);
  }
);

export const authApi = {
  login: async (credentials: LoginRequest): Promise<LoginResponse> => {
    const { data } = await apiClient.post<LoginResponse>('/api/v1/auth/login', credentials);
    return data;
  },
  
  register: async (userData: RegisterRequest): Promise<RegisterResponse> => {
    const { data } = await apiClient.post<RegisterResponse>('/api/v1/auth/register', userData);
    return data;
  },
  
  refresh: async (): Promise<TokenRefreshResponse> => {
    const { data } = await axios.post<TokenRefreshResponse>(
      `${API_BASE_URL}/api/v1/auth/refresh`,
      {},
      { withCredentials: true }
    );
    return data;
  },
};

export const portfolioApi = {
  list: async (page = 1, pageSize = 20): Promise<PortfolioListResponse> => {
    const { data } = await apiClient.get<PortfolioListResponse>('/api/v1/portfolios', {
      params: { page, pageSize },
    });
    return data;
  },
  
  addWallet: async (wallet: AddWalletRequest): Promise<AddWalletResponse> => {
    const { data } = await apiClient.post<AddWalletResponse>('/api/v1/portfolios', wallet);
    return data;
  },
};

export const walletApi = {
  getWallet: async (address: string, chain?: Chain): Promise<WalletView> => {
    const { data } = await apiClient.get<WalletView>(`/api/v1/wallets/${address}`, {
      params: chain ? { chain } : {},
    });
    return data;
  },
  
  getTransactions: async (
    address: string,
    page = 1,
    pageSize = 50
  ): Promise<TransactionPage> => {
    const { data } = await apiClient.get<TransactionPage>(
      `/api/v1/wallets/${address}/transactions`,
      {
        params: { page, pageSize },
      }
    );
    return data;
  },
  
  exportWallet: async (address: string, chain: Chain, format: 'csv' | 'json' = 'json'): Promise<Blob> => {
    const { data } = await apiClient.get(`/api/v1/export/wallet/${address}`, {
      params: { chain, format },
      responseType: 'blob',
    });
    return data;
  },
};

export const tokenApi = {
  getTopHolders: async (tokenAddress: string, chain?: Chain): Promise<TopHoldersResponse> => {
    const { data } = await apiClient.get<TopHoldersResponse>(
      `/api/v1/tokens/${tokenAddress}/holders`,
      {
        params: chain ? { chain } : {},
      }
    );
    return data;
  },
};

export { apiClient };
