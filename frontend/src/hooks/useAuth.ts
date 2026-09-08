import { useAuth } from '../context/AuthContext';
import { useNavigate, useLocation } from 'react-router-dom';
import { useEffect } from 'react';
import { getAuthReturnTo } from '../lib/auth-return';

/**
 * Hook to access auth state
 * Alias for useAuth for consistency with naming convention
 */
export { useAuth };

/**
 * Hook to get the current user
 * Returns null if not authenticated
 */
export function useUser() {
  const { user } = useAuth();
  return user;
}

/**
 * Hook to check if user is authenticated
 */
export function useIsAuthenticated() {
  const { isAuthenticated } = useAuth();
  return isAuthenticated;
}

/**
 * Hook that forces authentication
 * Redirects to login if not authenticated
 */
export function useRequireAuth() {
  const { isAuthenticated, isLoading } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      // Store the current location for redirect after login
      sessionStorage.setItem('auth_return_to', getAuthReturnTo(location));
      navigate('/login', { replace: true, state: { from: location } });
    }
  }, [isAuthenticated, isLoading, navigate, location]);

  return { isAuthenticated, isLoading };
}
