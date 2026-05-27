FROM node:22.22-alpine AS builder  
WORKDIR /app  
COPY package.json ./  
RUN npm install  
COPY . .  
RUN npm run build  
  
FROM nginx:alpine
COPY ./nginx.conf /etc/nginx/conf.d/default.conf  
COPY --from=builder /app/dist /usr/share/nginx/html
COPY ./env-config.template.js /usr/share/nginx/html/env-config.template.js
RUN printf '%s\n' '#!/bin/sh' 'envsubst '\''$FRONTEND_URL $PROJECT_VERSION'\'' < /usr/share/nginx/html/env-config.template.js > /usr/share/nginx/html/env-config.js' > /docker-entrypoint.d/40-env-config.sh \
 && chmod +x /docker-entrypoint.d/40-env-config.sh
EXPOSE 80
