/**
 * Authentication hook for managing user session state.
 * Copyright (c) 2025 aezizhu. All rights reserved.
 * Author: aezizhu
 * Repository: github.com/aezizhu/million-dollar-hunter
 */
'use client';

import { useState, useEffect } from 'react';
import { authApi } from '@/lib/api';
import type { LoginRequest, RegisterRequest } from '@/types';

export function useAuth() {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('accessToken');
      setIsAuthenticated(!!token);
    }
    setIsLoading(false);
  }, []);

  const login = async (credentials: LoginRequest) => {
    try {
      const response = await authApi.login(credentials);
      if (typeof window !== 'undefined') {
        localStorage.setItem('accessToken', response.accessToken);
        if (response.refreshToken) {
          localStorage.setItem('refreshToken', response.refreshToken);
        }
      }
      setIsAuthenticated(true);
      return { success: true };
    } catch (error) {
      const err = error as { response?: { data?: { message?: string } } };
      return {
        success: false,
        error: err.response?.data?.message || 'Login failed',
      };
    }
  };

  const register = async (userData: RegisterRequest) => {
    try {
      await authApi.register(userData);
      return { success: true };
    } catch (error) {
      const err = error as { response?: { data?: { message?: string } } };
      return {
        success: false,
        error: err.response?.data?.message || 'Registration failed',
      };
    }
  };

  const logout = () => {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      setIsAuthenticated(false);
      window.location.href = '/auth/login';
    }
  };

  return {
    isAuthenticated,
    isLoading,
    login,
    register,
    logout,
  };
}
