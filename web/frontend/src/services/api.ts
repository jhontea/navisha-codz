import axios, { AxiosError, AxiosInstance, InternalAxiosRequestConfig } from "axios";
import type {
  ApiResponse,
  AuthResponse,
  LeaderboardEntry,
  LeaderboardPeriod,
  LoginRequest,
  PaginatedResponse,
  Problem,
  RegisterRequest,
  Submission,
  SubmitRequest,
  SubmitResponse,
  TestResult,
  User,
} from "../types";

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api/v1";
const WS_BASE_URL = import.meta.env.VITE_WS_BASE_URL || "ws://localhost:8080/ws";

// --- Axios Instance ---

const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    "Content-Type": "application/json",
  },
});

// Request interceptor — attach token
api.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem("access_token");
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor — handle 401 + token refresh
let isRefreshing = false;
let failedQueue: Array<{ resolve: (token: string) => void; reject: (err: unknown) => void }> = [];

const processQueue = (error: unknown, token: string | null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token!);
    }
  });
  failedQueue = [];
};

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiResponse<unknown>>) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({
            resolve: (token: string) => {
              if (originalRequest.headers) {
                originalRequest.headers.Authorization = `Bearer ${token}`;
              }
              resolve(api(originalRequest));
            },
            reject,
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const refreshToken = localStorage.getItem("refresh_token");
        if (!refreshToken) {
          throw new Error("No refresh token");
        }

        const { data } = await axios.post<ApiResponse<AuthResponse>>(`${API_BASE_URL}/auth/refresh`, {
          refresh_token: refreshToken,
        });

        const { access_token, refresh_token: newRefresh } = data.data;
        localStorage.setItem("access_token", access_token);
        localStorage.setItem("refresh_token", newRefresh);

        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${access_token}`;
        }

        processQueue(null, access_token);
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        window.location.href = "/login";
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

// --- Auth API ---

export const authApi = {
  login: (credentials: LoginRequest) =>
    api.post<ApiResponse<AuthResponse>>("/auth/login", credentials).then((r) => r.data),

  register: (data: RegisterRequest) =>
    api.post<ApiResponse<AuthResponse>>("/auth/register", data).then((r) => r.data),

  refresh: (refreshToken: string) =>
    api
      .post<ApiResponse<AuthResponse>>("/auth/refresh", { refresh_token: refreshToken })
      .then((r) => r.data),

  logout: () => api.post("/auth/logout").then((r) => r.data),

  getProfile: () => api.get<ApiResponse<User>>("/auth/profile").then((r) => r.data),
};

// --- Problem API ---

export const problemApi = {
  list: (params?: {
    page?: number;
    page_size?: number;
    difficulty?: string;
    category?: string;
    search?: string;
    status?: string;
  }) => api.get<ApiResponse<PaginatedResponse<Problem>>>("/problems", { params }).then((r) => r.data),

  getById: (id: string) =>
    api.get<ApiResponse<Problem>>(`/problems/${id}`).then((r) => r.data),

  getBySlug: (slug: string) =>
    api.get<ApiResponse<Problem>>(`/problems/slug/${slug}`).then((r) => r.data),

  submit: (data: SubmitRequest) =>
    api.post<ApiResponse<SubmitResponse>>("/problems/submit", data).then((r) => r.data),
};

// --- Submission API ---

export const submissionApi = {
  getStatus: (id: string) =>
    api.get<ApiResponse<Submission>>(`/submissions/${id}`).then((r) => r.data),

  getHistory: (params?: { page?: number; page_size?: number; problem_id?: string }) =>
    api
      .get<ApiResponse<PaginatedResponse<Submission>>>("/submissions", { params })
      .then((r) => r.data),

  getTestResults: (id: string) =>
    api.get<ApiResponse<TestResult[]>>(`/submissions/${id}/results`).then((r) => r.data),
};

// --- Leaderboard API ---

export const leaderboardApi = {
  get: (period: LeaderboardPeriod = "all-time", page = 1, pageSize = 50) =>
    api
      .get<ApiResponse<PaginatedResponse<LeaderboardEntry>>>("/leaderboard", {
        params: { period, page, page_size: pageSize },
      })
      .then((r) => r.data),
};

// --- Hint API ---

export const hintApi = {
  reveal: (problemId: string, hintId: string) =>
    api
      .post<ApiResponse<{ hint: { id: string; content: string }; penalty: number }>>(
        `/problems/${problemId}/hints/${hintId}/reveal`
      )
      .then((r) => r.data),

  getRevealed: (problemId: string) =>
    api
      .get<ApiResponse<Array<{ id: string; level: number; content: string }>>>(
        `/problems/${problemId}/hints`
      )
      .then((r) => r.data),
};

export { api, API_BASE_URL, WS_BASE_URL };
