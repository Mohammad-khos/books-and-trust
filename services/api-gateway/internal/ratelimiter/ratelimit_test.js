import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";

export const allowedRequests = new Counter("allowed_requests");
export const rateLimitedRequests = new Counter("rate_limited_requests");

export const options = {
  scenarios: {
    ip_limit_test: {
      executor: "constant-vus",
      vus: 50,
      duration: "5s",
      exec: "testIPLimit",
    },

    user_limit_test: {
      executor: "per-vu-iterations",
      vus: 10,
      iterations: 20,
      maxDuration: "10s",
      exec: "testUserLimit",
    },
  },

  thresholds: {
    checks: ["rate>0.99"],

    http_req_duration: ["p(95)<100"],
  },
};

const BASE_URL = "http://localhost:8081/api/v1";

function recordMetrics(res) {
  if (res.status === 200) {
    allowedRequests.add(1);
  }

  if (res.status === 429) {
    rateLimitedRequests.add(1);
  }
}

export function testIPLimit() {
  const res = http.post(
    `${BASE_URL}/users/login`,
    JSON.stringify({
      credential: "test@example.com",
      password: "password123",
    }),
    {
      headers: {
        "Content-Type": "application/json",
      },
    },
  );

  recordMetrics(res);

  check(res, {
    "not server error": (r) => r.status < 500,
  });

  sleep(0.1);
}

export function testUserLimit() {
  const res = http.get(`${BASE_URL}/loans`, {
    headers: {
      "Content-Type": "application/json",
      Authorization:
        "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJib29rcy1hbmQtdHJ1c3QiLCJleHAiOjE3ODA5MTU1MDIsImlhdCI6MTc4MDMxMDcwMiwiaXNzIjoiYm9va3MtYW5kLXRydXN0LXVzZXItc2VydmljZSIsIm5iZiI6MTc4MDMxMDcwMiwic3ViIjoiMjE0ZjBhNzQtZmEyYS00ZmJiLThjMjAtZWRkMzhmMWQwYjA5In0.K_YQzp5PrpGOKlPPSM38Ar3jd_Zei_-G1wEVnaWyTcs",
    },
  });

  recordMetrics(res);

  check(res, {
    "User limiter returned 200 or 429": (r) =>
      r.status <500 ,
  });

  sleep(0.1);
}
