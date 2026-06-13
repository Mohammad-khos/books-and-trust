import http from "k6/http";
import { check, sleep } from "k6";

export const options = {
  vus: 10,
  duration: "10s",
};

export default function () {
  const res = http.post(
    "http://localhost:8081/api/v1/loans",
    JSON.stringify({
        "owner_id" : "4b42f0db-e3c0-492c-ac55-b7609eb7a14e",
        "book_name" :"test book name"
    }),
    {
      headers: {
        "Content-Type": "application/json",
        Authorization:
          "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJib29rcy1hbmQtdHJ1c3QiLCJleHAiOjE3ODEyNjkzMDYsImlhdCI6MTc4MDY2NDUwNiwiaXNzIjoiYm9va3MtYW5kLXRydXN0LXVzZXItc2VydmljZSIsIm5iZiI6MTc4MDY2NDUwNiwic3ViIjoiNGI0MmYwZGItZTNjMC00OTJjLWFjNTUtYjc2MDllYjdhMTRlIn0.rmwuyG1nybsvSBNTj91n5pJVbJv48XSHU3x-dgG3Nqs",
      },
    },
  );

  check(res, {
    "rate limited (429)": (r) => r.status === 429,
    "bad request (400)" : (r) => r.status === 400,
    "circuit breaker tripped (503)": (r) => r.status === 503,
    "internal server error (500)": (r) => r.status === 500,
  });

  sleep(0.1);
}
