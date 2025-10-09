export interface User {
  userId: string;
  email: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
}

export interface RegisterResponse {
  userId: string;
  message: string;
}

export interface TokenRefreshResponse {
  accessToken: string;
}

export interface PortfolioSummary {
  id: string;
  address: string;
  chain: string;
  netWorthUsd: number;
  nickname?: string;
}

export interface PortfolioListResponse {
  items: PortfolioSummary[];
  page: number;
  pageSize: number;
  total: number;
}

export interface AddWalletRequest {
  address: string;
  chain: string;
  nickname?: string;
}

export interface AddWalletResponse {
  jobId: string;
  status: 'queued' | 'running';
}

export interface Asset {
  tokenAddress: string;
  symbol: string;
  balance: number;
  usdValue: number;
  name?: string;
}

export interface HistoryPoint {
  time: string;
  value: number;
}

export interface WalletView {
  address: string;
  chain: string;
  assets: Asset[];
  history: HistoryPoint[];
  netWorthUsd?: number;
  nickname?: string;
}

export interface Transaction {
  timestamp: string;
  type: 'SEND' | 'RECEIVE' | 'SWAP' | 'APPROVE' | 'MINT' | 'BURN';
  from: string;
  to: string;
  symbol: string;
  amount: number;
  usdValue: number;
  txHash: string;
}

export interface TransactionPage {
  items: Transaction[];
  page: number;
  pageSize: number;
  total: number;
}

export interface Holder {
  address: string;
  balance: number;
  percent: number;
}

export interface TopHoldersResponse {
  token: string;
  holders: Holder[];
}

export interface ApiError {
  error: string;
  message: string;
  details?: Record<string, unknown>;
}

export type Chain = 'ethereum' | 'bsc' | 'polygon' | 'arbitrum' | 'optimism' | 'solana';
