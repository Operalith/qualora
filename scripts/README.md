# Scripts

Developer automation will live here.

Scripts should be safe to run locally, documented, and avoid printing secrets.

Current scripts:

- `smoke.py`: end-to-end browser and API smoke test driver. It prints JSON and HTML report URLs, validates HTML report export, and validates screenshot evidence download.
- `mock-api/server.js`: deterministic local API used by `make smoke`.
- `demo-web/server.js`: deterministic local frontend used by browser smoke tests.
