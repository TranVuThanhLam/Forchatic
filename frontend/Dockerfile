FROM node:18

WORKDIR /app

# Cài serve và dependencies
COPY package.json package-lock.json ./
RUN npm install -g serve && npm install

# Copy toàn bộ mã nguồn và build
COPY . .
COPY .env.prod .env
RUN npm run build

EXPOSE 3000

# Lệnh chạy khi container khởi động
CMD ["serve", "-s", "build", "-l", "3000"]
