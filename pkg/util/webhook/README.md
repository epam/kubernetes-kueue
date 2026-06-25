# pkg/util/webhook/

Webhook utility helpers for consistent webhook handler implementation.

## Key Functions

- `FromObject(obj runtime.Object, gvk schema.GroupVersionKind) *unstructured.Unstructured` — convert typed object to unstructured for webhook handlers
- `ExtractList(req admission.Request) ([]runtime.Object, error)` — parse objects from a webhook request
- `WithAnnotations(resp admission.Response, annotations map[string]string) admission.Response` — add annotations to a webhook response

## Usage

Each webhook handler (`pkg/webhooks/`) uses these helpers to reduce boilerplate in request/response handling.
