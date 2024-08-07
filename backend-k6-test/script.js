import http from 'k6/http';
import {sleep} from 'k6';

export const options = {
  // A number specifying the number of VUs to run concurrently.
  vus: 4,
  // A string specifying the total duration of the test run.
  duration: '5s',

  // Uncomment this section to enable the use of Browser API in your tests.
  //
  // See https://grafana.com/docs/k6/latest/using-k6-browser/running-browser-tests/ to learn more
  // about using Browser API in your test scripts.
  //
  // scenarios: {
  //   // The scenario name appears in the result summary, tags, and so on.
  //   // You can give the scenario any name, as long as each name in the script is unique.
  //   ui: {
  //     // Executor is a mandatory parameter for browser-based tests.
  //     // Shared iterations in this case tells k6 to reuse VUs to execute iterations.
  //     //
  //     // See https://grafana.com/docs/k6/latest/using-k6/scenarios/executors/ for other executor types.
  //     executor: 'shared-iterations',
  //     options: {
  //       browser: {
  //         // This is a mandatory parameter that instructs k6 to launch and
  //         // connect to a chromium-based browser, and use it to run UI-based
  //         // tests.
  //         type: 'chromium',
  //       },
  //     },
  //   },
  // }
};

// The function that defines VU logic.
export default function() {
  http.get('http://localhost:8080/auth/test');
  http.get('http://localhost:8080/open/pages');
  http.get('localhost:8080/open/login');
  http.get('localhost:8080/admin/pages');
  sleep(1);
}
