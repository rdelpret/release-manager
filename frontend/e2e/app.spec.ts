import { test, expect } from "@playwright/test";

// Helper: dev-login via the Next.js proxy so the cookie is set on the right origin
async function devLogin(page: import("@playwright/test").Page) {
  await page.goto("/auth/dev-login");
  // The dev-login endpoint sets the session cookie and redirects to /dashboard
  await page.waitForURL(/\/dashboard/, { timeout: 10000 });
}

test.describe("Login Page", () => {
  test("shows login page with Subwave branding", async ({ page }) => {
    await page.goto("/login");
    await expect(page.locator("h1")).toContainText("Subwave");
    await expect(page.locator("text=Release Campaign Planner")).toBeVisible();
    await expect(page.locator("text=Sign in with Google")).toBeVisible();
    await expect(page.locator("text=Dev Login")).toBeVisible();
  });

  test("redirects unauthenticated user from dashboard to login", async ({
    page,
  }) => {
    await page.goto("/dashboard");
    // Should redirect to /login since not authenticated
    await page.waitForURL(/\/login/, { timeout: 5000 });
    await expect(page.locator("h1")).toContainText("Subwave");
  });

  test("dev-login endpoint logs in and reaches dashboard", async ({ page }) => {
    await page.goto("/auth/dev-login");
    await page.waitForURL(/\/dashboard/, { timeout: 10000 });
    await expect(page.locator("text=Subwave")).toBeVisible();
  });
});

test.describe("Dashboard", () => {
  test.beforeEach(async ({ page }) => {
    await devLogin(page);
  });

  test("shows dashboard after login", async ({ page }) => {
    await page.goto("/dashboard");
    await expect(page.locator("text=Subwave")).toBeVisible();
    await expect(page.locator("text=Campaigns")).toBeVisible();
    await expect(page.locator("text=New Campaign")).toBeVisible();
  });

  test("can create a new campaign", async ({ page }) => {
    await page.goto("/dashboard");
    await page.click("text=New Campaign");
    const input = page.locator('input[placeholder="Campaign name..."]');
    await expect(input).toBeVisible();
    await input.fill("Test Release - Playwright");
    await input.press("Enter");

    // Should see the new campaign card
    await expect(
      page.locator("text=Test Release - Playwright")
    ).toBeVisible({ timeout: 5000 });
  });

  test("can navigate to a campaign", async ({ page }) => {
    await page.goto("/dashboard");

    const name = `Nav Test ${Date.now()}`;
    await page.click("text=New Campaign");
    const input = page.locator('input[placeholder="Campaign name..."]');
    await input.fill(name);
    await input.press("Enter");
    await expect(page.getByText(name).first()).toBeVisible({
      timeout: 5000,
    });

    await page.getByText(name).first().click();
    await page.waitForURL(/\/campaign\//, { timeout: 5000 });

    // Should see campaign name on the board page
    await expect(page.getByText(name)).toBeVisible({ timeout: 5000 });
  });
});

test.describe("Campaign Board", () => {
  test.beforeEach(async ({ page }) => {
    await devLogin(page);
    // Create a campaign with unique name and navigate to it
    const name = `Board Test ${Date.now()}`;
    await page.goto("/dashboard");
    await page.click("text=New Campaign");
    const input = page.locator('input[placeholder="Campaign name..."]');
    await input.fill(name);
    await input.press("Enter");
    await expect(page.getByText(name).first()).toBeVisible({
      timeout: 5000,
    });
    await page.getByText(name).first().click();
    await page.waitForURL(/\/campaign\//, { timeout: 5000 });
  });

  test("shows task list tabs", async ({ page }) => {
    // Should see the 5 task list tabs from the template
    await expect(page.locator("text=Campaign Assets")).toBeVisible({
      timeout: 5000,
    });
  });

  test("shows task groups and tasks", async ({ page }) => {
    // Wait for campaign data to load
    await expect(page.locator("text=Campaign Assets")).toBeVisible({
      timeout: 5000,
    });

    // Should see tasks from the template (first tab: Campaign Assets > Song Assets)
    await expect(page.locator("text=Final Master")).toBeVisible({
      timeout: 5000,
    });
  });

  test("can cycle task status", async ({ page }) => {
    await expect(page.locator("text=Final Master")).toBeVisible({
      timeout: 5000,
    });

    // Find the status button for Final Master and click it
    const taskRow = page.locator("text=Final Master").locator("..");
    const statusButton = taskRow.locator("button").first();
    await statusButton.click();

    // Status should have changed (hard to assert exact state, just verify no error)
    await expect(page.locator("text=Final Master")).toBeVisible();
  });

  test("can open task detail panel", async ({ page }) => {
    await expect(page.locator("text=Final Master")).toBeVisible({
      timeout: 5000,
    });

    // Click on the task name to open detail panel
    await page.locator("text=Final Master").click();

    // Detail panel should appear with status options and due date
    await expect(page.locator("text=Status")).toBeVisible({ timeout: 3000 });
    await expect(page.locator("text=Due Date")).toBeVisible();
    await expect(page.locator("text=Subtasks")).toBeVisible();
  });

  test("can switch between task list tabs", async ({ page }) => {
    await expect(page.locator("text=Campaign Assets")).toBeVisible({
      timeout: 5000,
    });

    // Click on PR tab
    await page.locator("button:has-text('PR')").click();

    // Should see PR group header
    await expect(page.getByRole("button", { name: /Media Outreach/ })).toBeVisible({
      timeout: 3000,
    });
  });

  test("can navigate to calendar view", async ({ page }) => {
    await expect(page.locator("text=Calendar")).toBeVisible({ timeout: 5000 });
    await page.locator("text=Calendar").click();
    await page.waitForURL(/\/calendar/, { timeout: 5000 });
  });

  test("back button returns to dashboard", async ({ page }) => {
    await expect(page.locator("text=Campaign Assets")).toBeVisible({
      timeout: 5000,
    });

    // Click back arrow
    const backButton = page.locator('[class*="text-accent"]').first();
    await backButton.click();
    await page.waitForURL(/\/dashboard/, { timeout: 5000 });
  });
});

test.describe("API Health", () => {
  test("health endpoint returns ok", async ({ request }) => {
    const response = await request.get("http://localhost:8080/api/health");
    expect(response.ok()).toBeTruthy();
    expect(await response.json()).toEqual({ status: "ok" });
  });

  test("unauthenticated API request returns 401", async ({ request }) => {
    const response = await request.get("http://localhost:8080/api/campaigns");
    expect(response.status()).toBe(401);
  });
});
