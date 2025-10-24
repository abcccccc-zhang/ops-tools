FROM node:20.12.2 AS builder
RUN npm config set registry http://registry.npm.taobao.org
WORKDIR /app
COPY . /app/
RUN npm install \
    && npm run build

FROM 10.10.10.109/nginx/nginxinc:nginx-unprivileged
ENV TZ=Asia/Shanghai
COPY --from=builder /app/dist /usr/share/nginx/html/dist
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
