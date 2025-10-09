import { renderHook, act } from '@testing-library/react';
import { useAuth } from '../useAuth';
import { authApi } from '@/lib/api';

jest.mock('@/lib/api');

describe('useAuth', () => {
  const mockSetItem = jest.fn();
  const mockRemoveItem = jest.fn();

  beforeEach(() => {
    jest.clearAllMocks();
    Storage.prototype.setItem = mockSetItem;
    Storage.prototype.removeItem = mockRemoveItem;
    Storage.prototype.getItem = jest.fn();
  });

  it('initializes with no authentication', () => {
    const { result } = renderHook(() => useAuth());
    
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('initializes as authenticated if token exists', () => {
    Storage.prototype.getItem = jest.fn(() => 'test-token');
    
    const { result } = renderHook(() => useAuth());
    
    expect(result.current.isAuthenticated).toBe(true);
  });

  it('logs in successfully', async () => {
    const mockLogin = authApi.login as jest.MockedFunction<typeof authApi.login>;
    mockLogin.mockResolvedValue({
      accessToken: 'test-token',
      refreshToken: 'refresh-token',
    });

    const { result } = renderHook(() => useAuth());
    
    let loginResult;
    await act(async () => {
      loginResult = await result.current.login({
        email: 'test@example.com',
        password: 'password',
      });
    });

    expect(loginResult).toEqual({ success: true });
    expect(result.current.isAuthenticated).toBe(true);
    expect(mockSetItem).toHaveBeenCalledWith('accessToken', 'test-token');
  });

  it('handles login failure', async () => {
    const mockLogin = authApi.login as jest.MockedFunction<typeof authApi.login>;
    mockLogin.mockRejectedValue({
      response: { data: { message: 'Invalid credentials' } },
    });

    const { result } = renderHook(() => useAuth());
    
    let loginResult;
    await act(async () => {
      loginResult = await result.current.login({
        email: 'test@example.com',
        password: 'wrong',
      });
    });

    expect(loginResult).toEqual({
      success: false,
      error: 'Invalid credentials',
    });
    expect(result.current.isAuthenticated).toBe(false);
  });

  it('calls logout and clears token', () => {
    Storage.prototype.getItem = jest.fn(() => 'test-token');
    
    const { result } = renderHook(() => useAuth());
    
    result.current.logout();
    
    expect(mockRemoveItem).toHaveBeenCalledWith('accessToken');
  });
});
