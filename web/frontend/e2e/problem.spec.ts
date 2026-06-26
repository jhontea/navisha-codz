import { test, expect } from "@playwright/test";
import { mockProblems, apiResponse } from "./mock-data";

const API_HOST = "http://localhost:8080";

function setupMocks(context: any) {
  return context.route("**/*", async (route, request) => {
    const url = request.url();
    if (!url.includes(API_HOST)) {
      // Not an API request — let it through
      return route.continue();
    }

    // Problem by slug
    if (url.includes("/problems/slug/")) {
      const slug = url.split("/").pop();
      const problem = mockProblems.find((p) => p.slug === slug);
      if (problem) {
        return route.fulfill({
          status: 200,
          contentType: "application/json",
          body: JSON.stringify(apiResponse(problem)),
        });
      }
      return route.fulfill({
        status: 404,
        contentType: "application/json",
        body: JSON.stringify({ success: false, error: "Not found" }),
      });
    }

    // Problems list
    if (url.includes("/problems")) {
      return route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(
          apiResponse({ items: mockProblems, total: 2, page: 1, page_size: 20, total_pages: 1 })
        ),
      });
    }

    // Auth endpoints
    if (url.includes("/auth/login") || url.includes("/auth/register")) {
      return route.fulfill({
        status: 200,
        contentType: "application/json",
        body: JSON.stringify(
          apiResponse({
            user: { id: "1", username: "test", email: "test@test.com", role: "user", score: 0, rank: 0, rating: 0, streak_days: 0, max_streak_days: 0, created_at: "", updated_at: "" },
            access_token: "mock-token",
            refresh_token: "mock-refresh",
            expires_in: 3600,
          })
        ),
      });
    }

    route.continue();
  });
}

test.describe("Problem Detail", () => {
  test.beforeEach(async ({ page, context }) => {
    await page.addInitScript(() => {
      localStorage.setItem("auth-storage", JSON.stringify({
        state: {
          user: { id: "user-1", username: "testuser", email: "test@example.com", role: "user", score: 500, rank: 10, rating: 1200, streak_days: 5, max_streak_days: 10, created_at: "", updated_at: "" },
          accessToken: "mock-access-token-12345",
          refreshToken: "mock-refresh-token-67890",
          isAuthenticated: true,
        },
        version: 0,
      }));
      localStorage.setItem("access_token", "mock-access-token-12345");
      localStorage.setItem("refresh_token", "mock-refresh-token-67890");
    });

    await setupMocks(context);
  });

  test("navigasi ke problem detail via URL langsung", async ({ page }) => {
    await page.goto("/problems/two-sum");
    await page.waitForLoadState("networkidle");

    await expect(page.locator("h1").filter({ hasText: "Two Sum" })).toBeVisible({ timeout: 15000 });
    await expect(page.getByText("Easy").first()).toBeVisible();
    await expect(page.getByText("100 points")).toBeVisible();
    await expect(page.getByText("1000ms")).toBeVisible();
    await expect(page.getByText("256MB")).toBeVisible();
    await expect(page.getByText(/Given an array of integers nums/i)).toBeVisible();
    await expect(page.getByText(/Example 1/)).toBeVisible();
    await expect(page.getByText(/Constraints:/)).toBeVisible();

    await page.screenshot({ path: "e2e/screenshots/problem-two-sum.png", fullPage: false });
  });

  test("navigasi via problems list lalu klik problem", async ({ page }) => {
    await page.goto("/problems");
    await page.waitForLoadState("networkidle");

    await page.waitForSelector('a[href="/problems/two-sum"]', { timeout: 10000 });
    await page.click('a[href="/problems/two-sum"]');

    await expect(page).toHaveURL(/\/problems\/two-sum/);
    await expect(page.locator("h1").filter({ hasText: "Two Sum" })).toBeVisible({ timeout: 15000 });
  });

  test("navigasi antar problem", async ({ page }) => {
    await page.goto("/problems");
    await page.waitForLoadState("networkidle");

    await page.waitForSelector('a[href="/problems/reverse-linked-list"]', { timeout: 10000 });
    await page.click('a[href="/problems/reverse-linked-list"]');

    await expect(page).toHaveURL(/\/problems\/reverse-linked-list/);
    await expect(page.locator("h1").filter({ hasText: "Reverse Linked List" })).toBeVisible({ timeout: 15000 });
  });

  test("halaman error untuk problem tidak ditemukan", async ({ page }) => {
    await page.goto("/problems/non-existent");
    await page.waitForLoadState("networkidle");

    await expect(page.getByText(/Failed to load problem/i)).toBeVisible({ timeout: 10000 });
    await expect(page.getByRole("link", { name: /Back to problems/i })).toBeVisible();
  });

  test("code editor muncul di problem detail", async ({ page }) => {
    await page.goto("/problems/two-sum");
    await page.waitForLoadState("networkidle");

    await expect(page.locator("h1").filter({ hasText: "Two Sum" })).toBeVisible({ timeout: 15000 });

    // Monaco editor loads from CDN — butuh waktu
    const editor = page.locator(".monaco-editor");
    await expect(editor).toBeVisible({ timeout: 30000 });

    await expect(page.getByRole("button", { name: /Submit solution/i })).toBeVisible();
    await expect(page.getByRole("button", { name: /Reset to template/i })).toBeVisible();

    await page.screenshot({ path: "e2e/screenshots/problem-editor.png", fullPage: false });
  });
});
