import { test, expect } from "@playwright/test";
import { mockApiRoutes, mockProblems } from "./mock-data";

test.describe("Homepage", () => {
  test("memuat homepage dengan hero section, stats, dan featured problems", async ({
    page,
  }) => {
    // Mock API
    await mockApiRoutes(page);

    // Navigate to homepage
    await page.goto("/");

    // Hero section — heading utama
    const heading = page.getByRole("heading", {
      name: /code\.? compete\.? conquer/i,
    });
    await expect(heading).toBeVisible({ timeout: 10000 });

    // Subtitle
    await expect(
      page.getByText(/Sharpen your algorithmic thinking/i)
    ).toBeVisible();

    // Stats section
    const statsSection = page.locator("section").filter({ hasText: "Problems" });
    await expect(statsSection.getByText("500+")).toBeVisible();
    await expect(statsSection.getByText("10K+")).toBeVisible();
    await expect(statsSection.getByText("1M+")).toBeVisible();

    // Featured Problems section
    const featured = page.getByRole("heading", {
      name: /Featured Problems/i,
    });
    await expect(featured).toBeVisible();

    // Featured problem cards muncul
    const firstProblem = page.getByRole("link", { name: /Two Sum/ });
    await expect(firstProblem).toBeVisible({ timeout: 10000 });
    await expect(
      page.getByRole("link", { name: /Reverse Linked List/ })
    ).toBeVisible();

    // Difficulty badge
    await expect(page.getByText("easy").first()).toBeVisible();
    await expect(page.getByText("medium").first()).toBeVisible();

    // Tombol CTA
    await expect(
      page.getByRole("link", { name: /Get Started Free/i })
    ).toBeVisible();
    await expect(
      page.getByRole("link", { name: /Browse Problems/i })
    ).toBeVisible();

    // Features section
    await expect(
      page.getByRole("heading", { name: /Why CodeChallenge/i })
    ).toBeVisible();
    await expect(page.getByText(/Algorithm Mastery/)).toBeVisible();
    await expect(page.getByText(/Real-time Feedback/)).toBeVisible();
    await expect(page.getByText(/Compete & Climb/)).toBeVisible();

    // Quick Start section
    await expect(
      page.getByRole("heading", { name: /Ready to Start/i })
    ).toBeVisible();

    // Screenshot
    await page.screenshot({
      path: "e2e/screenshots/homepage.png",
      fullPage: true,
    });
  });

  test("menampilkan problem list dengan benar", async ({ page }) => {
    await mockApiRoutes(page);
    await page.goto("/");

    // Klik "Browse Problems" — karena belum login, akan redirect ke /problems yang protected
    // Tapi kita test bahwa featured problems muncul di homepage
    const viewAll = page.getByRole("link", { name: /View all/i });
    await expect(viewAll).toBeVisible();

    // Featured problems memiliki title, difficulty, points
    const problemCards = page.locator('a[href^="/problems/"]');
    const count = await problemCards.count();
    expect(count).toBeGreaterThanOrEqual(2);

    // Verifikasi data di card
    await expect(
      page.locator("text=100 pts").first()
    ).toBeVisible();
    await expect(
      page.locator("text=200 pts").first()
    ).toBeVisible();
  });

  test("link navigasi berfungsi", async ({ page }) => {
    await mockApiRoutes(page);
    await page.goto("/");

    // Klik "Start Solving" — karena belum login akan redirect ke login
    // Tapi kita bisa verifikasi link href-nya
    const startSolving = page.getByRole("link", { name: /Start Solving/i });
    // "Start Solving" hanya muncul untuk authenticated user, jadi "Get Started Free" untuk anonymous
    const getStarted = page.getByRole("link", { name: /Get Started Free/i });
    await expect(getStarted).toBeVisible();
    await expect(getStarted).toHaveAttribute("href", "/register");

    const browseProblems = page.getByRole("link", { name: /Browse Problems/i });
    await expect(browseProblems).toBeVisible();
    await expect(browseProblems).toHaveAttribute("href", "/problems");
  });
});
