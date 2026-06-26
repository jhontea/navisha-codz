#!/usr/bin/env node
/**
 * Bundle Analysis Script
 * Generates a visual report of bundle size and chunk breakdown
 * Usage: node scripts/analyze.js
 */

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const ROOT = path.resolve(__dirname, "..");
const DIST = path.join(ROOT, "dist");

console.log("📊 Starting bundle analysis...\n");

// Check if dist exists
if (!fs.existsSync(DIST)) {
  console.log("❌ dist/ directory not found. Running build first...\n");
  try {
    execSync("npm run build", { cwd: ROOT, stdio: "inherit" });
  } catch (e) {
    console.error("Build failed:", e.message);
    process.exit(1);
  }
}

// Analyze dist folder
function getDirSize(dirPath) {
  let size = 0;
  const files = fs.readdirSync(dirPath);
  for (const file of files) {
    const filePath = path.join(dirPath, file);
    const stat = fs.statSync(filePath);
    if (stat.isDirectory()) {
      size += getDirSize(filePath);
    } else {
      size += stat.size;
    }
  }
  return size;
}

function formatBytes(bytes) {
  if (bytes === 0) return "0 B";
  const k = 1024;
  const sizes = ["B", "KB", "MB", "GB"];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + " " + sizes[i];
}

const assetsDir = path.join(DIST, "assets");
if (fs.existsSync(assetsDir)) {
  console.log("📦 Bundle Breakdown:");
  console.log("─".repeat(60));

  const files = fs.readdirSync(assetsDir).filter((f) => f.endsWith(".js") || f.endsWith(".css"));

  let totalSize = 0;
  const fileData = [];

  for (const file of files) {
    const filePath = path.join(assetsDir, file);
    const stat = fs.statSync(filePath);
    totalSize += stat.size;
    fileData.push({ name: file, size: stat.size });
  }

  // Sort by size descending
  fileData.sort((a, b) => b.size - a.size);

  for (const { name, size } of fileData) {
    const percentage = ((size / totalSize) * 100).toFixed(1);
    const bar = "█".repeat(Math.ceil(percentage / 2));
    console.log(
      `  ${name.padEnd(40)} ${formatBytes(size).padStart(10)} ${bar} ${percentage}%`
    );
  }

  console.log("─".repeat(60));
  console.log(`  ${"TOTAL".padEnd(40)} ${formatBytes(totalSize).padStart(10)}`);
  console.log("\n✅ Bundle analysis complete!");
  console.log(`📄 Visual report: ${path.join(DIST, "bundle-stats.html")}`);
} else {
  console.log("⚠️  No assets directory found in dist/");
}
