/** User profile */
export interface User {
  id: string;
  username: string;
  email: string;
  avatar_url?: string;
  role: "user" | "admin";
  score: number;
  rank: number;
  rating: number;
  streak_days: number;
  max_streak_days: number;
  badges?: Badge[];
  achievements?: Achievement[];
  created_at: string;
  updated_at: string;
}

/** Badge model */
export interface Badge {
  id: string;
  name: string;
  emoji: string;
  description: string;
  unlocked_at?: string;
}

/** Achievement model */
export interface Achievement {
  id: string;
  name: string;
  description: string;
  icon: string;
  unlocked_at?: string;
  progress?: number;
  max_progress?: number;
}

/** DP Visualization types */
export interface DPCell {
  row: number;
  col: number;
  value: number | string;
  color: string;
  formula?: string;
  dependencies?: Array<{ row: number; col: number }>;
  is_base_case?: boolean;
  is_result?: boolean;
  is_backtrack?: boolean;
  explanation?: string;
}

export interface DPStep {
  step_number: number;
  description: string;
  cells_modified: DPCell[];
  active_cells: Array<{ row: number; col: number }>;
  highlight_cells: Array<{ row: number; col: number }>;
  current_value?: number | string;
  formula?: string;
  memo_state?: MemoFrame[];
  complexity_now?: string;
}

export interface MemoFrame {
  function: string;
  args: unknown;
  result?: unknown;
  depth: number;
  cached: boolean;
}

export interface DPDimensions {
  rows: number;
  cols: number;
  row_labels: string[];
  col_labels: string[];
}

export interface DPTableResult {
  problem_id: number;
  title: string;
  description: string;
  table: DPCell[][];
  steps: DPStep[];
  dimensions: DPDimensions;
  approach: "memoization" | "tabulation";
  time_complexity: string;
  space_complexity: string;
  memo_stack?: MemoFrame[];
  backtrack_path?: Array<{ row: number; col: number }>;
  optimal_cells?: Array<{ row: number; col: number }>;
  state_colors?: Record<string, string>;
}

/** Difficulty level */
export type Difficulty = "easy" | "medium" | "hard";

/** Problem category */
export type Category =
  | "arrays"
  | "strings"
  | "linked-lists"
  | "trees"
  | "graphs"
  | "dynamic-programming"
  | "sorting"
  | "math"
  | "greedy"
  | "backtracking";

/** Problem model */
export interface Problem {
  id: string;
  title: string;
  slug: string;
  description: string;
  difficulty: Difficulty;
  category: Category;
  tags: string[];
  time_limit_ms: number;
  memory_limit_mb: number;
  points: number;
  solved_count: number;
  attempt_count: number;
  examples: ProblemExample[];
  constraints: string[];
  function_template: string;
  test_cases: TestCase[];
  hints: Hint[];
  created_at: string;
  updated_at: string;
}

/** Example input/output for a problem */
export interface ProblemExample {
  id: string;
  input: string;
  output: string;
  explanation?: string;
}

/** Test case */
export interface TestCase {
  id: string;
  input: string;
  expected_output: string;
  is_sample: boolean;
  description?: string;
}

/** Submission status */
export type SubmissionStatus =
  | "pending"
  | "running"
  | "accepted"
  | "wrong_answer"
  | "time_limit_exceeded"
  | "memory_limit_exceeded"
  | "runtime_error"
  | "compilation_error";

/** Submission model */
export interface Submission {
  id: string;
  problem_id: string;
  user_id: string;
  code: string;
  language: string;
  status: SubmissionStatus;
  score: number;
  execution_time_ms: number;
  memory_used_kb: number;
  test_results: TestResult[];
  error_message?: string;
  created_at: string;
}

/** Individual test result */
export interface TestResult {
  test_case_id: string;
  status: SubmissionStatus;
  actual_output?: string;
  expected_output: string;
  execution_time_ms: number;
  memory_used_kb: number;
  error_message?: string;
}

/** Hint model */
export interface Hint {
  id: string;
  problem_id: string;
  level: number;
  content: string;
  penalty_points: number;
  is_revealed: boolean;
}

/** Leaderboard entry */
export interface LeaderboardEntry {
  rank: number;
  user_id: string;
  username: string;
  avatar_url?: string;
  score: number;
  problems_solved: number;
  total_submissions: number;
  accuracy: number;
}

/** Leaderboard period */
export type LeaderboardPeriod = "weekly" | "monthly" | "all-time";

// --- API Response Types ---

export interface ApiResponse<T> {
  success: boolean;
  data: T;
  message?: string;
  error?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
  total_pages: number;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  username: string;
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface SubmitRequest {
  problem_id: string;
  code: string;
  language: string;
}

export interface SubmitResponse {
  submission_id: string;
  status: SubmissionStatus;
}

// --- WebSocket Message Types ---

export type WsMessageType =
  | "submission_update"
  | "test_result"
  | "leaderboard_update"
  | "connected"
  | "error";

export interface WsMessage<T = unknown> {
  type: WsMessageType;
  payload: T;
  timestamp: string;
}

export interface SubmissionUpdatePayload {
  submission_id: string;
  status: SubmissionStatus;
  progress: number;
  completed_tests: number;
  total_tests: number;
}

export interface TestResultPayload {
  submission_id: string;
  test_result: TestResult;
}

export interface LeaderboardUpdatePayload {
  period: LeaderboardPeriod;
  entries: LeaderboardEntry[];
}
