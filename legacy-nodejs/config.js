// require("dotenv").config();
import "dotenv/config";

const baseConfig = {
  web_url: process.env.WEB_URL,
  socket: {
    topik: {
      status: process.env.TOPIC_STATUS || "status",
    },
    url: process.env.SOCKET_URL,
  },
};

export default baseConfig;
