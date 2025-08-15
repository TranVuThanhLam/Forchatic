const http = require("http");
const WebSocket = require("ws");
const path = require("path");
const fs = require("fs");

const server = http.createServer((req, res) => {
  // Serve file tĩnh index.html
  if (req.url === "/" || req.url === "/index.html") {
    const filePath = path.join(__dirname, "public", "index.html");
    fs.createReadStream(filePath).pipe(res);
  } else {
    res.writeHead(404);
    res.end("Not Found");
  }
});

// Tạo WebSocket server
const wss = new WebSocket.Server({ server });

wss.on("connection", (ws) => {
  console.log("Client connected");
  ws.send("Chào từ server WebSocket!");

  ws.on("message", (message) => {
    console.log("Nhận từ client:", message.toString());
    // Gửi lại cho tất cả client
    wss.clients.forEach((client) => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(`Echo: ${message}`);
      }
    });
  });

  ws.on("close", () => console.log("Client disconnected"));
});

// Server HTTP + WS chạy trên 1 port
const PORT = 3000;
server.listen(PORT, () => {
  console.log(`Server chạy tại http://0.0.0.0:${PORT}`);
});
