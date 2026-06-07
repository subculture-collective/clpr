import axios, { AxiosError } from 'axios';
import type { AxiosResponse, InternalAxiosRequestConfig } from 'axios';

const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";

export const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    "Content-Type": "application/json",
  },
  withCredentials: true, // Important: Send cookies with requests
});

// Type-safe queue item for failed requests during token refresh
interface QueuedRequest {
  resolve: (value: AxiosResponse) => void;
  reject: (reason: AxiosError) => void;
}

// Flag to prevent multiple refresh attempts
let isRefreshing = false;
let failedQueue: QueuedRequest[] = [];

// Store CSRF token from response headers
let csrfToken: string | null = null;

type UnauthorizedHandler = (error: AxiosError) => void;

let unauthorizedHandler: UnauthorizedHandler | null = null;

export const setUnauthorizedHandler = (
  handler: UnauthorizedHandler | null
): void => {
  unauthorizedHandler = handler;
};

const notifyUnauthorized = (error: AxiosError): void => {
  if (unauthorizedHandler) {
    unauthorizedHandler(error);
  }
};

const isUnauthorizedError = (error: AxiosError): boolean =>
  error.response?.status === 401;

// Request interceptor to add CSRF token to state-changing requests
apiClient.interceptors.request.use(
  (config) => {
    if (typeof FormData !== 'undefined' && config.data instanceof FormData) {
      // Let Axios/the browser set multipart/form-data with the generated boundary.
      if (typeof config.headers.delete === 'function') {
        config.headers.delete('Content-Type');
      } else {
        delete config.headers['Content-Type'];
      }
    }

    // Only add CSRF token for state-changing methods
    if (config.method && ['post', 'put', 'delete', 'patch'].includes(config.method.toLowerCase())) {
      if (csrfToken) {
        config.headers['X-CSRF-Token'] = csrfToken;
      }
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

const processQueue = (
  error: AxiosError | null,
  token: AxiosResponse | null = null
) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else if (token) {
      prom.resolve(token);
    } else {
      prom.reject(new AxiosError('Token refresh failed'));
    }
  });
  failedQueue = [];
};

// Response interceptor to handle token refresh on 401
apiClient.interceptors.response.use(
  (response) => {
    // Extract CSRF token from response headers if present
    const token = response.headers['x-csrf-token'];
    if (token) {
      csrfToken = token;
    }
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    if (isUnauthorizedError(error) && originalRequest?._retry) {
      notifyUnauthorized(error);
      return Promise.reject(error);
    }

    // If error is 401 and we haven't tried to refresh yet
    if (
      isUnauthorizedError(error) &&
      originalRequest &&
      !originalRequest._retry
    ) {
      // Check if this is a refresh token request failing
      if (originalRequest.url === "/auth/refresh") {
        // Refresh token is invalid, logout user
        isRefreshing = false;
        processQueue(error, null);
        notifyUnauthorized(error);
        // Let the AuthContext handle the logout
        return Promise.reject(error);
      }

      if (isRefreshing) {
        // If already refreshing, queue this request
        return new Promise<AxiosResponse>((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        })
          .then(() => {
            return apiClient(originalRequest);
          })
          .catch((err) => {
            return Promise.reject(err);
          });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        // Try to refresh the token
        const response = await apiClient.post('/auth/refresh');

        // Token refreshed successfully (new tokens are in cookies)
        isRefreshing = false;
        processQueue(null, response);

        // Retry the original request
        return apiClient(originalRequest);
      } catch (refreshError) {
        // Refresh failed
        isRefreshing = false;
        const refreshAxiosError = refreshError as AxiosError;
        processQueue(refreshAxiosError, null);
        if (isUnauthorizedError(refreshAxiosError)) {
          notifyUnauthorized(refreshAxiosError);
        }
        return Promise.reject(refreshError);
      }
    }

    return Promise.reject(error);
  }
);

export default apiClient;
