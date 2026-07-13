# Control Plane

Planned Go service for Qualora's API and orchestration layer.

Responsibilities:

- Project configuration.
- Test run lifecycle.
- Policy validation.
- Worker job scheduling.
- Metadata persistence.
- Report and finding access.

The control plane should remain small in the MVP and delegate long-running checks to workers as those workers are implemented.
