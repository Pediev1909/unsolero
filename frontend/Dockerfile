FROM node:22-alpine@sha256:c610fcdfb1d5b4740dd70c284ed3cb16bb857e0f7166196e36a5501df7a3aa32 AS dependencies

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM dependencies AS development
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

FROM dependencies AS build
COPY . .
RUN npm run build

FROM nginx:1.29-alpine@sha256:5616878291a2eed594aee8db4dade5878cf7edcb475e59193904b198d9b830de AS production
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY nginx-static-security.conf /etc/nginx/rigmark-static-security.conf
COPY --from=build /app/dist /usr/share/nginx/html
RUN sed -i 's/listen 80;/listen 8080;/' /etc/nginx/conf.d/default.conf && \
    chown -R nginx:nginx /var/cache/nginx /var/run /usr/share/nginx/html /etc/nginx/conf.d
USER nginx
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD wget -qO- http://127.0.0.1:8080/ || exit 1

FROM production AS staging
USER root
RUN apk add --no-cache openssl
COPY nginx-staging-tls.conf /etc/nginx/conf.d/staging-tls.conf
COPY nginx-staging-main.conf /etc/nginx/nginx.conf
COPY staging-entrypoint.sh /usr/local/bin/staging-entrypoint
RUN chmod 0555 /usr/local/bin/staging-entrypoint
USER nginx
EXPOSE 8443
ENTRYPOINT ["/usr/local/bin/staging-entrypoint"]
