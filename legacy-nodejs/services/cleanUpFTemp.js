import fs from "fs";

function cleanUp(files) {
  console.log("Cleaning up files...");
  files.forEach((file) => {
    try {
      if (fs.existsSync(file)) {
        fs.unlinkSync(file);
        console.log(`Deleted file: ${file}`);
      }
    } catch (error) {
      console.error(`Error deleting file: ${file}`, error);
    }
  });
  console.log("Clean up done!");
}

export { cleanUp };
