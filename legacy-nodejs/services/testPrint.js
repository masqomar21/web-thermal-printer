import escpos from "escpos";
import escposUSB from "escpos-usb";

let device;
let printer;

escpos.USB = escposUSB;

function initPrinter() {
  try {
    device = new escpos.USB();
    printer = new escpos.Printer(device);
  } catch (error) {
    console.error("Printer initialization failed, Retrying in 1s...");
    setTimeout(initPrinter, 1000); // Prevents infinite recursion
  }
}

initPrinter();

async function testPrint() {
  try {
    if (!device || !printer) {
      initPrinter();
      await new Promise((resolve) => setTimeout(resolve, 1000));
      return testPrint();
    }

    return new Promise((resolve) => {
      device.open((err) => {
        if (err) {
          console.error("Error opening device:", err);
          setTimeout(() => testPrint().then(resolve), 1000);
          return;
        }

        printer
          .font("a")
          .align("ct")
          .style("bu")
          .size(1, 1)
          .text("printer ready!")
          .newLine()
          .newLine()
          .newLine()
          .cut()
          .close();

        resolve(true);
      });
    });
  } catch (error) {
    // console.error("Error printing ticket:", error);
    initPrinter();
    await new Promise((resolve) => setTimeout(resolve, 1000));
    return testPrint();
  }
}

export { testPrint };
