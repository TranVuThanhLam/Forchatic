FROM nginx:alpine

RUN rm /etc/nginx/conf.d/default.conf
COPY default.conf /etc/nginx/conf.d/default.conf

# Cài netcat để kiểm tra kết nối
RUN apk add --no-cache netcat-openbsd

# Thêm script chờ backend
COPY wait-for-backends.sh /wait-for-backends.sh
RUN chmod +x /wait-for-backends.sh

# Dùng script như entrypoint
CMD ["/wait-for-backends.sh"]

# EXPOSE 4000

