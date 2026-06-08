export async function onRequest(context) {
  const { request, env } = context;
  const url = new URL(request.url);
  
  // The base URL of your Go backend (Refinery)
  // You should set this as an environment variable in Cloudflare Pages dashboard
  const BACKEND_URL = env.BACKEND_URL || "http://localhost:8080"; 

  // Rewrite the URL to point to the backend
  const backendRequestUrl = request.url.replace(url.origin + "/api", BACKEND_URL);

  // Clone the request with the new URL
  const backendRequest = new Request(backendRequestUrl, request);

  // Fetch from the backend
  return fetch(backendRequest);
}
