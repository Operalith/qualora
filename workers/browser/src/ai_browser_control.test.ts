import test from "node:test";
import assert from "node:assert/strict";
import {
  evaluateAIBrowserPolicy,
  parseAIBrowserSuggestion,
  redactObservationValue,
  type AIBrowserObservation
} from "./ai_browser_control";

const observation: AIBrowserObservation = {
  current_url: "http://demo-web:8080/",
  current_path: "/",
  page_title: "Qualora Demo Web",
  visible_text_snippets: ["Self-hosted QA automation demo"],
  headings: ["Qualora Demo Web"],
  safe_candidate_links: [
    {
      text: "About",
      path: "/about",
      target_url: "http://demo-web:8080/about",
      same_origin: true,
      safety: "safe",
      selector_hint: 'a:text("About")'
    }
  ],
  candidate_buttons: [{ label: "Delete account", safety: "unsafe", selector_hint: "button#delete-account" }],
  forms: [
    {
      method: "GET",
      field_count: 1,
      password_field_count: 0,
      classification: "search",
      safety: "safe",
      selector_hint: "form#site-search",
      action_url: "http://demo-web:8080/search",
      fields: [{ name: "q", type: "search", label: "Search" }]
    },
    {
      method: "POST",
      field_count: 2,
      password_field_count: 1,
      classification: "unknown",
      safety: "unsafe",
      selector_hint: "form#login",
      action_url: "http://demo-web:8080/login",
      fields: [{ name: "password", type: "password", label: "Password" }]
    }
  ],
  console_error_count: 0,
  failed_request_count: 0,
  previous_steps: []
};

const policy = {
  sourceURL: "http://demo-web:8080/",
  frontendURL: "http://demo-web:8080",
  allowedHosts: ["demo-web"],
  sameOriginOnly: true,
  allowGetForms: false
};

test("parseAIBrowserSuggestion rejects invalid JSON", () => {
  const result = parseAIBrowserSuggestion("{nope");
  assert.equal(result.suggestion, null);
  assert.match(result.error, /valid JSON/);
});

test("evaluateAIBrowserPolicy approves observed safe links", () => {
  const parsed = parseAIBrowserSuggestion(
    JSON.stringify({
      rationale: "About is an observed safe same-origin link.",
      action: { type: "click_link", target_url: "/about", link_text: "About" },
      expected_result: "About page loads.",
      risk_assessment: "safe_navigation"
    })
  );
  const decision = evaluateAIBrowserPolicy({
    suggestion: parsed.suggestion,
    observation,
    policy,
    startURL: "http://demo-web:8080/",
    depth: 0,
    maxDepth: 2,
    visited: new Set(["http://demo-web:8080/"])
  });
  assert.equal(decision.decision, "approved");
  assert.equal(decision.targetURL, "http://demo-web:8080/about");
});

test("evaluateAIBrowserPolicy blocks destructive paths", () => {
  const parsed = parseAIBrowserSuggestion(
    JSON.stringify({
      rationale: "Try unsafe link.",
      action: { type: "goto", target_url: "/delete-account" },
      expected_result: "Account deleted.",
      risk_assessment: "unsafe"
    })
  );
  const decision = evaluateAIBrowserPolicy({
    suggestion: parsed.suggestion,
    observation,
    policy,
    startURL: "http://demo-web:8080/",
    depth: 0,
    maxDepth: 2,
    visited: new Set()
  });
  assert.equal(decision.decision, "blocked");
  assert.match(decision.reason, /destructive|mutating/);
});

test("evaluateAIBrowserPolicy blocks external URLs and sensitive query values", () => {
  const external = parseAIBrowserSuggestion(
    JSON.stringify({ rationale: "Leave site.", action: { type: "goto", target_url: "https://example.com/" } })
  );
  assert.equal(
    evaluateAIBrowserPolicy({ suggestion: external.suggestion, observation, policy, startURL: observation.current_url, depth: 0, maxDepth: 2, visited: new Set() }).decision,
    "blocked"
  );

  const token = parseAIBrowserSuggestion(
    JSON.stringify({ rationale: "Token URL.", action: { type: "goto", target_url: "/about?token=abc" } })
  );
  const tokenDecision = evaluateAIBrowserPolicy({ suggestion: token.suggestion, observation, policy, startURL: observation.current_url, depth: 0, maxDepth: 2, visited: new Set() });
  assert.equal(tokenDecision.decision, "blocked");
  assert.match(tokenDecision.reason, /sensitive query/);
});

test("redactObservationValue redacts obvious secret material", () => {
  const value = redactObservationValue("Authorization=Bearer abc token=secret password=hunter2");
  assert.doesNotMatch(value, /hunter2|secret|Bearer abc/);
  assert.match(value, /\[REDACTED\]/);
});

test("evaluateAIBrowserPolicy approves observed safe GET forms", () => {
  const parsed = parseAIBrowserSuggestion(
    JSON.stringify({
      rationale: "Submit the observed safe search form.",
      action: {
        type: "submit_safe_get_form",
        form_selector_hint: "form#site-search",
        field_values: { q: "demo" },
        label: "Search demo"
      },
      expected_result: "Search result page loads.",
      risk_assessment: "safe_navigation"
    })
  );
  const decision = evaluateAIBrowserPolicy({
    suggestion: parsed.suggestion,
    observation,
    policy: { ...policy, allowGetForms: false },
    startURL: observation.current_url,
    depth: 0,
    maxDepth: 2,
    visited: new Set()
  });
  assert.equal(decision.decision, "approved");
  assert.equal(decision.targetURL, "http://demo-web:8080/search?q=demo");

  const noLabel = parseAIBrowserSuggestion(
    JSON.stringify({
      rationale: "Submit the observed safe search form without a label.",
      action: {
        type: "submit_safe_get_form",
        form_selector_hint: "form#site-search",
        field_values: { q: "demo" }
      }
    })
  );
  const noLabelDecision = evaluateAIBrowserPolicy({
    suggestion: noLabel.suggestion,
    observation,
    policy: { ...policy, allowGetForms: false },
    startURL: observation.current_url,
    depth: 0,
    maxDepth: 2,
    visited: new Set()
  });
  assert.equal(noLabelDecision.decision, "approved");
  assert.equal(noLabelDecision.targetURL, "http://demo-web:8080/search?q=demo");
});

test("evaluateAIBrowserPolicy blocks unsafe or unobserved forms", () => {
  const unsafe = parseAIBrowserSuggestion(
    JSON.stringify({
      rationale: "Try a login form.",
      action: {
        type: "submit_safe_get_form",
        form_selector_hint: "form#login",
        field_values: { password: "demo" },
        label: "Login"
      }
    })
  );
  const unsafeDecision = evaluateAIBrowserPolicy({
    suggestion: unsafe.suggestion,
    observation,
    policy,
    startURL: observation.current_url,
    depth: 0,
    maxDepth: 2,
    visited: new Set()
  });
  assert.equal(unsafeDecision.decision, "blocked");
  assert.match(unsafeDecision.reason, /not classified safe|destructive|mutating|sensitive|not safe GET/);

  const unobserved = parseAIBrowserSuggestion(
    JSON.stringify({
      rationale: "Invent a form.",
      action: {
        type: "submit_safe_get_form",
        form_selector_hint: "form#made-up",
        field_values: { q: "demo" },
        label: "Search"
      }
    })
  );
  const unobservedDecision = evaluateAIBrowserPolicy({
    suggestion: unobserved.suggestion,
    observation,
    policy,
    startURL: observation.current_url,
    depth: 0,
    maxDepth: 2,
    visited: new Set()
  });
  assert.equal(unobservedDecision.decision, "blocked");
  assert.match(unobservedDecision.reason, /observed safe form/);
});
