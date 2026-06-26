import type { Problem, ApiResponse, PaginatedResponse } from "../src/types/index";

/** Fake problem data for E2E tests */
export const mockProblems: Problem[] = [
  {
    id: "1",
    title: "Two Sum",
    slug: "two-sum",
    description:
      "Given an array of integers nums and an integer target, return indices of the two numbers that add up to target.",
    difficulty: "easy",
    category: "arrays",
    tags: ["array", "hash-map"],
    time_limit_ms: 1000,
    memory_limit_mb: 256,
    points: 100,
    solved_count: 5000,
    attempt_count: 8000,
    examples: [
      {
        id: "ex1",
        input: "nums = [2,7,11,15], target = 9",
        output: "[0, 1]",
        explanation: "Because nums[0] + nums[1] == 9, we return [0, 1].",
      },
    ],
    constraints: ["2 <= nums.length <= 10^4", "-10^9 <= nums[i] <= 10^9"],
    function_template:
      "func twoSum(nums []int, target int) []int {\n\t// Write your code here\n}",
    test_cases: [
      { id: "tc1", input: "[2,7,11,15]\n9", expected_output: "[0,1]", is_sample: true },
      { id: "tc2", input: "[3,2,4]\n6", expected_output: "[1,2]", is_sample: true },
    ],
    hints: [
      {
        id: "h1",
        problem_id: "1",
        level: 1,
        content: "Try using a hash map to store seen values.",
        penalty_points: 10,
        is_revealed: false,
      },
    ],
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
  },
  {
    id: "2",
    title: "Reverse Linked List",
    slug: "reverse-linked-list",
    description:
      "Given the head of a singly linked list, reverse the list and return the reversed list.",
    difficulty: "medium",
    category: "linked-lists",
    tags: ["linked-list", "recursion"],
    time_limit_ms: 1000,
    memory_limit_mb: 256,
    points: 200,
    solved_count: 3200,
    attempt_count: 6000,
    examples: [
      {
        id: "ex1",
        input: "head = [1,2,3,4,5]",
        output: "[5,4,3,2,1]",
      },
    ],
    constraints: ["0 <= nodes <= 5000"],
    function_template:
      "func reverseList(head *ListNode) *ListNode {\n\t// Write your code here\n}",
    test_cases: [
      { id: "tc1", input: "[1,2,3,4,5]", expected_output: "[5,4,3,2,1]", is_sample: true },
    ],
    hints: [
      {
        id: "h2",
        problem_id: "2",
        level: 1,
        content: "Use two pointers: prev and curr.",
        penalty_points: 10,
        is_revealed: false,
      },
    ],
    created_at: "2025-01-02T00:00:00Z",
    updated_at: "2025-01-02T00:00:00Z",
  },
];

/** Generate API response wrapper */
export function apiResponse(data) {
  return { success: true, data };
}

/** Paginated response helper */
export function paginated(items, total?) {
  return {
    items,
    total: total ?? items.length,
    page: 1,
    page_size: 20,
    total_pages: 1,
  };
}

/** Auth response mock */
export const mockAuthResponse = {
  user: {
    id: "user-1",
    username: "testuser",
    email: "test@example.com",
    role: "user",
    score: 500,
    rank: 10,
    rating: 1200,
    streak_days: 5,
    max_streak_days: 10,
    created_at: "2025-01-01T00:00:00Z",
    updated_at: "2025-01-01T00:00:00Z",
  },
  access_token: "mock-access-token-12345",
  refresh_token: "mock-refresh-token-67890",
  expires_in: 3600,
};

/** Set up localStorage auth state */
export async function setupAuth(page) {
  await page.addInitScript(() => {
    localStorage.setItem(
      "auth-storage",
      JSON.stringify({
        state: {
          user: {
            id: "user-1",
            username: "testuser",
            email: "test@example.com",
            role: "user",
            score: 500,
            rank: 10,
            rating: 1200,
            streak_days: 5,
            max_streak_days: 10,
            created_at: "2025-01-01T00:00:00Z",
            updated_at: "2025-01-01T00:00:00Z",
          },
          accessToken: "mock-access-token-12345",
          refreshToken: "mock-refresh-token-67890",
          isAuthenticated: true,
        },
        version: 0,
      })
    );
    localStorage.setItem("access_token", "mock-access-token-12345");
    localStorage.setItem("refresh_token", "mock-refresh-token-67890");
  });
}

/** Mock all API routes on the page */
export async function mockApiRoutes(page) {
  const API_BASE = "http://localhost:8080/api/v1";

  // Problems list — with or without query params
  await page.route(/\/api\/v1\/problems(\?.*)?$/, (route) => {
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(apiResponse(paginated(mockProblems))),
    });
  });

  // Problem by slug: /api/v1/problems/slug/:slug
  await page.route(/\/api\/v1\/problems\/slug\/.+$/, (route) => {
    const url = route.request().url();
    const slug = url.split("/").pop();
    const problem = mockProblems.find((p) => p.slug === slug);
    if (problem) {
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(apiResponse(problem)),
      });
    } else {
      route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ success: false, error: "Not found" }),
      });
    }
  });

  // Problem by ID: /api/v1/problems/:id (numeric)
  await page.route(/\/api\/v1\/problems\/\d+$/, (route) => {
    const url = route.request().url();
    const id = url.split("/").pop();
    const problem = mockProblems.find((p) => p.id === id);
    if (problem) {
      route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(apiResponse(problem)),
      });
    } else {
      route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ success: false, error: "Not found" }),
      });
    }
  });

  // Auth login
  await page.route("**/api/v1/auth/login", (route) => {
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(apiResponse(mockAuthResponse)),
    });
  });

  // Auth register
  await page.route("**/api/v1/auth/register", (route) => {
    route.fulfill({
      status: 200,
      contentType: "application/json",
      body: JSON.stringify(apiResponse(mockAuthResponse)),
    });
  });
}
