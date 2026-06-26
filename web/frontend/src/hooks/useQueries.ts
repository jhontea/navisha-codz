import { useQuery, useMutation, useQueryClient } from "react-query";
import { problemApi, submissionApi, leaderboardApi, hintApi } from "../services/api";
import type { Problem, Submission, LeaderboardEntry, LeaderboardPeriod, TestResult, PaginatedResponse } from "../types";

// --- Query Key Factory ---
export const queryKeys = {
  problems: {
    all: ["problems"] as const,
    list: (params: Record<string, unknown>) => ["problems", "list", params] as const,
    detail: (id: string) => ["problems", "detail", id] as const,
    bySlug: (slug: string) => ["problems", "slug", slug] as const,
  },
  submissions: {
    all: ["submissions"] as const,
    detail: (id: string) => ["submissions", "detail", id] as const,
    history: (params: Record<string, unknown>) => ["submissions", "history", params] as const,
    testResults: (id: string) => ["submissions", "test-results", id] as const,
  },
  leaderboard: {
    all: ["leaderboard"] as const,
    list: (period: LeaderboardPeriod) => ["leaderboard", "list", period] as const,
  },
  hints: {
    all: ["hints"] as const,
    forProblem: (problemId: string) => ["hints", "problem", problemId] as const,
  },
  profile: {
    all: ["profile"] as const,
  },
};

// --- Problem Queries ---
export function useProblemsList(params: {
  page?: number;
  page_size?: number;
  difficulty?: string;
  category?: string;
  search?: string;
}) {
  return useQuery<PaginatedResponse<Problem>>(
    queryKeys.problems.list(params),
    () => problemApi.list(params).then((r) => r.data),
    {
      staleTime: 60000, // 1 minute
    }
  );
}

export function useProblemBySlug(slug: string) {
  return useQuery<Problem>(
    queryKeys.problems.bySlug(slug),
    () => problemApi.getBySlug(slug).then((r) => r.data),
    {
      staleTime: 5 * 60 * 1000, // 5 minutes
      enabled: !!slug,
    }
  );
}

// --- Submission Queries with Optimistic Mutations ---
export function useSubmissionHistory(params: { page?: number; page_size?: number; problem_id?: string }) {
  return useQuery<PaginatedResponse<Submission>>(
    queryKeys.submissions.history(params),
    () => submissionApi.getHistory(params).then((r) => r.data),
    {
      staleTime: 30000,
      refetchInterval: 5000, // Poll every 5s for active submissions
    }
  );
}

export function useSubmissionDetail(id: string) {
  return useQuery<Submission>(
    queryKeys.submissions.detail(id),
    () => submissionApi.getStatus(id).then((r) => r.data),
    {
      enabled: !!id,
      refetchInterval: (data) => {
        if (data?.status === "pending" || data?.status === "running") return 2000;
        return false;
      },
    }
  );
}

export function useTestResults(submissionId: string) {
  return useQuery<TestResult[]>(
    queryKeys.submissions.testResults(submissionId),
    () => submissionApi.getTestResults(submissionId).then((r) => r.data),
    {
      enabled: !!submissionId,
      staleTime: 30000,
    }
  );
}

// --- Leaderboard Queries ---
export function useLeaderboard(period: LeaderboardPeriod = "all-time", page = 1, pageSize = 50) {
  return useQuery<PaginatedResponse<LeaderboardEntry>>(
    queryKeys.leaderboard.list(period),
    () => leaderboardApi.get(period, page, pageSize).then((r) => r.data),
    {
      staleTime: 60000,
      refetchInterval: 30000, // Refetch every 30s
    }
  );
}

// --- Hint Queries ---
export function useHints(problemId: string) {
  return useQuery<Array<{ id: string; level: number; content: string }>>(
    queryKeys.hints.forProblem(problemId),
    () => hintApi.getRevealed(problemId).then((r) => r.data),
    {
      enabled: !!problemId,
      staleTime: 5 * 60 * 1000,
    }
  );
}

// --- Optimistic Mutations ---
export function useSubmitSolution() {
  const queryClient = useQueryClient();

  return useMutation(
    (data: { problem_id: string; code: string; language: string }) =>
      problemApi.submit(data).then((r) => r.data),
    {
      onSuccess: (data) => {
        // Invalidate submission history
        queryClient.invalidateQueries(queryKeys.submissions.all);
        // Optimistically add to history
        queryClient.setQueryData(
          queryKeys.submissions.detail(data.submission_id),
          {
            id: data.submission_id,
            status: data.status,
          }
        );
      },
    }
  );
}

export function useRevealHint() {
  const queryClient = useQueryClient();

  return useMutation(
    ({ problemId, hintId }: { problemId: string; hintId: string }) =>
      hintApi.reveal(problemId, hintId).then((r) => r.data),
    {
      onSuccess: (_, { problemId }) => {
        queryClient.invalidateQueries(queryKeys.hints.forProblem(problemId));
      },
    }
  );
}
