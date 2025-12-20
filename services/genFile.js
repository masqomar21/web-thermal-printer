import fs from "fs";
import path from "path";
import puppeteer from "puppeteer";
import { fileURLToPath } from "url";
import baseConfig from "../config.js";

// Create __dirname equivalent for ESM
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

async function generateScreenshot(data) {
  const pagePath = path.resolve(__dirname, "../View/index.html");
  const html = fs.readFileSync(pagePath, "utf8");

  const pathTemp = path.resolve(__dirname, "../temp");
  if (!fs.existsSync(pathTemp)) {
    fs.mkdirSync(pathTemp);
  }

  const screenshotFilePath = path.resolve(
    __dirname,
    `../temp/ticket-${data.nomor_antrean}.png`
  );

  const logoPath = path.resolve(__dirname, "../images/logo.png");
  const logoBase64 = fs.readFileSync(logoPath, "base64");
  const logoDataURL = `data:image/png;base64,${logoBase64}`;

  const htmlContent = html
    // .replace("{{logo}}", `<img src="${logoDataURL}" alt="Logo">`)
    // .replace("{{instansi}}", data.instansi)
    .replace("{{loket}}", data.loket)
    .replace("{{nomor_antrean}}", data.nomor_antrean)
    .replace("{{date}}", data.tanggal)
    .replace("{{time}}", data.jam);
  // .replace("{{loket_name}}", data.loket)
  // .replace("{{web_url}}", baseConfig.web_url ?? "-");

  const browser = await puppeteer.launch();
  const page = await browser.newPage();

  await page.setContent(htmlContent);
  const container = await page.$(".container");

  if (container) {
    await container.screenshot({ path: screenshotFilePath });
    console.log(`Screenshot saved at: ${screenshotFilePath}`);
  } else {
    console.error("Container element not found!");
  }

  await browser.close();

  return screenshotFilePath;
}

export { generateScreenshot };
