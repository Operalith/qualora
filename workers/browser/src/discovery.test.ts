import assert from "node:assert/strict";
import test from "node:test";
import {
  buildDiscoveryFormFindings,
  buildDiscoveryPageFindings,
  classifyDiscoveryLink,
  normalizeDiscoveryURL,
  summarizeDiscoveryForm
} from "./discovery";

const policy = {
  sourceURL: "http://demo-web:8080/",
  frontendURL: "http://demo-web:8080/",
  allowedHosts: ["demo-web"],
  sameOriginOnly: true
};

test("normalizeDiscoveryURL strips fragments and redacts sensitive query values", () => {
  const normalized = normalizeDiscoveryURL("HTTP://DEMO-WEB:8080/dashboard?token=secret&page=1#top");
  assert.equal(normalized, "http://demo-web:8080/dashboard?page=1&token=%5BREDACTED%5D");
});

test("classifyDiscoveryLink allows safe same-origin links", () => {
  const decision = classifyDiscoveryLink("/about", "About", policy);
  assert.equal(decision.skipped, false);
  assert.equal(decision.sameOrigin, true);
  assert.equal(decision.normalizedURL, "http://demo-web:8080/about");
});

test("classifyDiscoveryLink skips external and unsafe links", () => {
  const external = classifyDiscoveryLink("https://example.com", "External", policy);
  assert.equal(external.skipped, true);
  assert.equal(external.skipReason, "external_link_skipped");

  const unsafe = classifyDiscoveryLink("/delete-account", "Delete account", policy);
  assert.equal(unsafe.skipped, true);
  assert.equal(unsafe.skipReason, "unsafe_link_skipped");
});

test("classifyDiscoveryLink skips unsupported schemes and downloads", () => {
  assert.equal(classifyDiscoveryLink("mailto:test@example.com", "Email", policy).skipReason, "unsupported_scheme");
  assert.equal(classifyDiscoveryLink("/report.pdf", "PDF", policy).skipReason, "non_html_resource");
});

test("summarizeDiscoveryForm captures password forms without submitting them", () => {
  const form = summarizeDiscoveryForm({
    form_name: "login",
    form_action: "/login",
    form_method: "POST",
    submit_button_count: 1,
    fields: [
      { field_name: "email", field_type: "email", placeholder: "Email", label: "Email", required: true },
      { field_name: "password", field_type: "password", placeholder: "", label: "", required: true }
    ]
  });
  assert.equal(form.classification, "password_form");
  assert.equal(form.password_field_count, 1);
  assert.equal(form.skipped_reason, "forms_are_not_submitted_by_discovery");
});

test("buildDiscovery finding helpers produce deterministic categories", () => {
  const pageFindings = buildDiscoveryPageFindings({
    url: "http://demo-web:8080/missing",
    statusCode: 404,
    loadError: "",
    bodyTextLength: 20,
    consoleErrorCount: 1,
    failedRequestCount: 1,
    evidenceIds: ["evidence-id"]
  });
  assert.deepEqual(
    pageFindings.map((finding) => finding.category),
    ["not_found", "console_error", "network_failure"]
  );

  const form = summarizeDiscoveryForm({
    form_name: "contact",
    form_action: "/contact",
    form_method: "get",
    submit_button_count: 1,
    fields: [{ field_name: "email", field_type: "email", placeholder: "", label: "", required: true }]
  });
  const formFindings = buildDiscoveryFormFindings(form, "http://demo-web:8080/", ["evidence-id"]);
  assert.equal(formFindings[0].category, "form_without_label");
});
