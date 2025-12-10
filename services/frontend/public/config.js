// Runtime configuration (injected by K8s ConfigMap in production)
// This file is loaded before the main app bundle
//
// Environments:
//   - local (docker-compose): Uses this default file
//   - develop (Kind): ConfigMap overwrites this file
//   - prod (EKS): ConfigMap overwrites this file
//
// ConfigMap example for K8s:
//   apiVersion: v1
//   kind: ConfigMap
//   metadata:
//     name: frontend-config
//   data:
//     config.js: |
//       window.__ENV__ = {
//         API_BASE_URL: "http://api.wealist.local"
//       };

window.__ENV__ = {
  // Default for local development (docker-compose)
  API_BASE_URL: "http://localhost"
};
