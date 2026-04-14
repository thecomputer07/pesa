# syntax=docker/dockerfile:1.7
# Dockerfile per il frontend Vite/React di Pesa.
# - target "dev"  → Vite dev server con HMR su :5173
# - target "prod" → static build servito da nginx su :80

FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

# ---- dev (Vite hot reload) ----
FROM deps AS dev
COPY . .
EXPOSE 5173
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

# ---- builder (prod) ----
FROM deps AS builder
COPY . .
RUN npm run build

# ---- prod runtime ----
FROM nginx:alpine AS prod
COPY --from=builder /app/dist /usr/share/nginx/html
# Fallback SPA: tutte le rotte → index.html
RUN printf 'server {\n  listen 80;\n  root /usr/share/nginx/html;\n  index index.html;\n  location / {\n    try_files $uri /index.html;\n  }\n}\n' > /etc/nginx/conf.d/default.conf
EXPOSE 80
