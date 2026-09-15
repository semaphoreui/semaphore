import axios from 'axios';

/**
 * Replaces the axios adapter so that no real HTTP requests are made.
 *
 * Usage:
 *   const http = mockAxios();          // in beforeEach
 *   http.respond((config) => ({ id: 1 }));
 *   ...
 *   http.restore();                    // in afterEach
 *
 * `http.requests` contains every request config seen (with `data` parsed
 * back from JSON when possible).
 */
export default function mockAxios() {
  const originalAdapter = axios.defaults.adapter;
  const requests = [];
  let responder = () => ({});

  axios.defaults.adapter = async (config) => {
    let { data } = config;
    if (typeof data === 'string') {
      try {
        data = JSON.parse(data);
      } catch (e) {
        // keep raw string
      }
    }
    requests.push({ ...config, data });

    const result = await responder({ ...config, data });
    return {
      data: result,
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
    };
  };

  return {
    requests,
    respond(fn) {
      responder = fn;
    },
    restore() {
      axios.defaults.adapter = originalAdapter;
    },
  };
}

/** Builds an error shaped like an axios HTTP error response. */
export function httpError(status, data, statusText = 'Error') {
  const err = new Error(`Request failed with status code ${status}`);
  err.response = { status, statusText, data };
  return err;
}
