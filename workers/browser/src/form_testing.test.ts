import test from "node:test";
import assert from "node:assert/strict";
import {
  buildSubmittedFormURL,
  classifyFormCandidate,
  formValuesSummary,
  type ExtractedFormCandidate,
  type FormTestPolicy
} from "./form_testing";

const policy: FormTestPolicy = {
  sourceURL: "http://demo-web:8080/",
  frontendURL: "http://demo-web:8080/",
  allowedHosts: ["demo-web"],
  sameOriginOnly: true,
  safeGetOnly: true
};

function form(overrides: Partial<ExtractedFormCandidate> = {}): ExtractedFormCandidate {
  return {
    selectorHint: "form#site-search",
    pageURL: "http://demo-web:8080/",
    actionURL: "http://demo-web:8080/search",
    method: "GET",
    label: "Site search",
    fieldCount: 1,
    passwordFieldCount: 0,
    fileFieldCount: 0,
    hiddenSensitiveFieldCount: 0,
    submitButtonCount: 1,
    fields: [
      {
        name: "q",
        type: "search",
        label: "Search",
        placeholder: "Search",
        required: false,
        hidden: false,
        options: []
      }
    ],
    ...overrides
  };
}

test("classifyFormCandidate allows same-origin safe GET search forms", () => {
  const result = classifyFormCandidate(form(), policy);
  assert.equal(result.safety, "safe");
  assert.equal(result.decision, "test");
  assert.equal(result.classification, "search");
  assert.deepEqual(result.testValues, { q: "demo" });
  assert.equal(buildSubmittedFormURL(result), "http://demo-web:8080/search?q=demo");

  const keyword = classifyFormCandidate(
    form({
      fields: [
        {
          name: "keywords",
          type: "search",
          label: "Keywords",
          placeholder: "Find docs",
          required: false,
          hidden: false,
          options: []
        }
      ]
    }),
    policy
  );
  assert.equal(keyword.safety, "safe");
  assert.deepEqual(keyword.testValues, { keywords: "demo" });
});

test("classifyFormCandidate skips mutating methods by default", () => {
  const result = classifyFormCandidate(form({ method: "POST", actionURL: "http://demo-web:8080/contact", label: "Contact support" }), policy);
  assert.equal(result.safety, "unsafe");
  assert.equal(result.decision, "skip");
  assert.equal(result.classification, "contact");
  assert.equal(result.skipReason, "form_method_not_safe");
});

test("classifyFormCandidate reuses allowed host validation", () => {
  const result = classifyFormCandidate(form({ actionURL: "https://example.com/search" }), policy);
  assert.equal(result.safety, "unsafe");
  assert.equal(result.decision, "skip");
  assert.equal(result.skipReason, "external_action_skipped");
});

test("classifyFormCandidate skips sensitive fields and dangerous workflows", () => {
  const password = classifyFormCandidate(
    form({
      actionURL: "http://demo-web:8080/login",
      label: "Login",
      passwordFieldCount: 1,
      fields: [
        {
          name: "password",
          type: "password",
          label: "Password",
          placeholder: "",
          required: true,
          hidden: false,
          options: []
        }
      ]
    }),
    policy
  );
  assert.equal(password.safety, "unsafe");
  assert.equal(password.classification, "password");
  assert.equal(password.skipReason, "sensitive_form_skipped");

  const deleteForm = classifyFormCandidate(form({ actionURL: "http://demo-web:8080/delete-account", label: "Delete account" }), policy);
  assert.equal(deleteForm.safety, "unsafe");
  assert.equal(deleteForm.classification, "destructive");
  assert.equal(deleteForm.skipReason, "unsafe_form_skipped");
});

test("classifyFormCandidate records explicit unsafe form purpose classes", () => {
  const newsletter = classifyFormCandidate(
    form({
      selectorHint: "form#newsletter-form",
      actionURL: "http://demo-web:8080/status",
      label: "Newsletter signup",
      fields: [
        {
          name: "newsletter_email",
          type: "email",
          label: "Newsletter email",
          placeholder: "person@example.test",
          required: true,
          hidden: false,
          options: []
        }
      ]
    }),
    policy
  );
  assert.equal(newsletter.classification, "newsletter");
  assert.equal(newsletter.safety, "unsafe");
  assert.equal(newsletter.skipReason, "sensitive_form_skipped");

  const payment = classifyFormCandidate(form({ actionURL: "http://demo-web:8080/checkout", label: "Checkout payment" }), policy);
  assert.equal(payment.classification, "payment");
  assert.equal(payment.safety, "unsafe");
});

test("formValuesSummary never stores raw submitted values", () => {
  const result = classifyFormCandidate(form(), policy);
  const summary = JSON.stringify(formValuesSummary(result));
  assert.doesNotMatch(summary, /demo/);
  assert.match(summary, /bounded_benign_value/);
});
