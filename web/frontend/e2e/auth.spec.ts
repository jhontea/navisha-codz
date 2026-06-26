import { test, expect } from "@playwright/test";
import { mockApiRoutes, mockAuthResponse } from "./mock-data";

test.describe("Authentication", () => {
  test("login form — fill dan submit sukses", async ({ page }) => {
    // Log API requests
    page.on("request", (req) => {
      if (req.url().includes("/api/")) console.log(`REQ: ${req.method()} ${req.url()}`);
    });
    page.on("response", (res) => {
      if (res.url().includes("/api/")) console.log(`RES: ${res.status()} ${res.url()}`);
    });
    page.on("console", (msg) => {
      if (msg.type() === "error") console.log(`[ERR] ${msg.text()}`);
    });

    await mockApiRoutes(page);

    await page.goto("/login");
    await page.waitForLoadState("networkidle");

    await page.locator("#email").fill("test@example.com");
    await page.locator("#password").fill("password123");

    await page.getByRole("button", { name: "Sign In", exact: true }).click();

    // Tunggu navigation
    await page.waitForTimeout(3000);
    console.log(`CURRENT URL: ${page.url()}`);

    await expect(page).toHaveURL(/\/problems/, { timeout: 15000 });
  });

  test("register mode — switch dan fill form", async ({ page }) => {
    page.on("request", (req) => {
      if (req.url().includes("/api/")) console.log(`REQ: ${req.method()} ${req.url()}`);
    });
    page.on("response", (res) => {
      if (res.url().includes("/api/")) console.log(`RES: ${res.status()} ${res.url()}`);
    });

    await mockApiRoutes(page);

    await page.goto("/login");
    await page.getByRole("button", { name: /Sign up/i }).click();

    await page.locator("#username").fill("testuser");
    await page.locator("#email").fill("test@example.com");
    await page.locator("#password").fill("password123");

    await page.getByRole("button", { name: /Create Account/i }).click();

    await page.waitForTimeout(3000);
    console.log(`CURRENT URL: ${page.url()}`);

    await expect(page).toHaveURL(/\/problems/, { timeout: 15000 });
  });
});
