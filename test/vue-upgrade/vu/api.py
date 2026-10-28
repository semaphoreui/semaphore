"""A logged-in Semaphore API session (cookies of one user) for the seed and the tests."""

import requests


class ApiError(Exception):
    def __init__(self, method, path, status, body):
        super().__init__(f"{method} {path} -> {status}: {body[:400]}")
        self.status = status
        self.body = body


class Api:
    def __init__(self, base_url, username, password):
        self.base = base_url.rstrip("/") + "/api"
        self.username = username
        self.http = requests.Session()
        self.call("POST", "/auth/login", {"auth": username, "password": password})

    def call(self, method, path, body=None, ok=(200, 201, 204)):
        resp = self.http.request(method, self.base + path, json=body, timeout=60,
                                 headers={"Accept": "application/json"})
        if resp.status_code not in ok:
            raise ApiError(method, path, resp.status_code, resp.text)
        text = resp.text.strip()
        if not text:
            return None
        try:
            return resp.json()
        except ValueError:
            return text

    def get(self, path):
        return self.call("GET", path)

    def post(self, path, body=None):
        return self.call("POST", path, body)

    def put(self, path, body):
        return self.call("PUT", path, body)

    def delete(self, path):
        return self.call("DELETE", path)

    def find(self, path, value, key="name"):
        """The item of the list at <path> whose <key> equals <value>, or None."""
        for item in self.get(path) or []:
            if item.get(key) == value:
                return item
        return None
