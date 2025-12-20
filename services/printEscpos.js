import escpos from "escpos";
import escposUSB from "escpos-usb";
import sharp from "sharp";
import path from "path";
import fs from "fs";
import { cleanUp } from "./cleanUpFTemp.js";
import { fileURLToPath } from "url";
import { io } from "socket.io-client";
import baseConfig from "../config.js";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const socket = io(baseConfig.socket.url);

escpos.USB = escposUSB;

let device;
let printer;

function initPrinter() {
  try {
    device = new escpos.USB();
    printer = new escpos.Printer(device);
  } catch (error) {
    console.error("Printer initialization failed:", error);
  }
}

initPrinter();

async function printTicket(pathToScreenshot) {
  console.log(pathToScreenshot);
  try {
    const pathTemp = path.resolve(__dirname, "../temp");
    if (!fs.existsSync(pathTemp)) {
      fs.mkdirSync(pathTemp);
    }
    const tempImagePath = path.join(
      __dirname,
      "../temp/resized-screenshot.png"
    );

    await sharp(pathToScreenshot).resize({ width: 600 }).toFile(tempImagePath);

    if (!device || !printer) {
      throw new Error("Printer not connected. Retrying on the next print.");
    }

    escpos.Image.load(tempImagePath, (image) => {
      device.open((err) => {
        if (err) {
          console.error("Error opening device:", err);
          socket.emit(baseConfig.socket.topik.status, {
            status: "printer_error",
            message: err.message,
          });
          initPrinter();
          return;
        }

        printer.align("CT").raster(image).newLine().newLine().cut().close();

        cleanUp([pathToScreenshot, tempImagePath]);
        socket.emit(baseConfig.socket.topik.status, { status: "printed" });
      });
    });
  } catch (error) {
    console.error("Error mencetak tiket:", error);
    socket.emit(baseConfig.socket.topik.status, {
      status: "error",
      message: error.message,
    });

    initPrinter();
  }
}
async function printTicketFromData(data) {
  console.log(data);
  try {
    if (!device || !printer) {
      throw new Error("Printer not connected. Retrying on the next print.");
    }

    device.open((err) => {
      if (err) {
        console.error("Error opening device:", err);
        socket.emit(baseConfig.socket.topik.status, {
          status: "printer_error",
          message: err.message,
        });
        initPrinter();
        return;
      }

      printer
        .font("A")
        .align("CT")
        .size(0, 1)
        .text("Loket : " + data.loket)
        .text("\n")
        .size(2, 2)
        .text(data.nomor_antrean)
        .text("\n")
        .size(0, 0.5)
        .text("Tanggal : " + data.tanggal)
        .text("Jam : " + data.jam)
        .cut()
        .close();

      // cleanUp([pathToScreenshot, tempImagePath]);
      socket.emit(baseConfig.socket.topik.status, { status: "printed" });
    });
  } catch (error) {
    console.error("Error mencetak tiket:", error);
    socket.emit(baseConfig.socket.topik.status, {
      status: "error",
      message: error.message,
    });

    initPrinter();
  }
}

export { printTicket, printTicketFromData };
