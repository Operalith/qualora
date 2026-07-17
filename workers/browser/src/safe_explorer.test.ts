import assert from "node:assert/strict";
import test from "node:test";
import {
  buildSafeExplorerActionFinding,
  classifySafeExplorerAction,
  markSafeExplorerDuplicate,
  normalizeSafeExplorerURL,
  type ExtractedSafeExplorerAction
} from "./safe_explorer";

const policy = {
  sourceURL: "http://demo-web:8080/",
  frontendURL: "http://demo-web:8080/",
  allowedHosts: ["demo-web"],
  sameOriginOnly: true,
  allowGetForms: false
};

function link(href: string, label = "Link"): ExtractedSafeExplorerAction {
  return {
    actionType: "link_navigation",
    label,
    text: label,
    selectorHint: `a:${label}`,
    href,
    targetURL: href,
    method: "GET"
  };
}

test("normalizeSafeExplorerURL strips fragments and redacts sensitive query values", () => {
  const normalized = normalizeSafeExplorerURL("HTTP://DEMO-WEB:8080/dashboard?token=secret&page=1#top");
  assert.equal(normalized, "http://demo-web:8080/dashboard?page=1&token=%5BREDACTED%5D");
});

test("classifySafeExplorerAction allows safe same-origin links", () => {
  const decision = classifySafeExplorerAction(link("/about", "About"), policy);
  assert.equal(decision.safety, "safe");
  assert.equal(decision.decision, "execute");
  assert.equal(decision.normalizedURL, "http://demo-web:8080/about");
});

test("classifySafeExplorerAction skips external and disallowed hosts", () => {
  const external = classifySafeExplorerAction(link("https://example.com", "External"), policy);
  assert.equal(external.decision, "skip");
  assert.equal(external.skipReason, "external_action_skipped");

  const sameOriginOnly = false;
  const disallowed = classifySafeExplorerAction(link("http://other.local/", "Other"), { ...policy, sameOriginOnly });
  assert.equal(disallowed.skipReason, "host_not_allowed");
});

test("classifySafeExplorerAction skips dangerous labels and sensitive query values", () => {
  const unsafe = classifySafeExplorerAction(link("/delete-account", "Delete account"), policy);
  assert.equal(unsafe.safety, "unsafe");
  assert.equal(unsafe.skipReason, "unsafe_action_skipped");

  const sensitive = classifySafeExplorerAction(link("/invite?token=secret", "Invite"), policy);
  assert.equal(sensitive.safety, "unsafe");
  assert.equal(sensitive.skipReason, "sensitive_query_skipped");
  assert.equal(sensitive.normalizedURL, "http://demo-web:8080/invite?token=%5BREDACTED%5D");
});

test("classifySafeExplorerAction skips unsafe and unsupported forms by default", () => {
  const getForm = classifySafeExplorerAction(
    {
      actionType: "form_get",
      label: "Search",
      text: "Search",
      selectorHint: "form#search",
      href: "",
      targetURL: "/search",
      method: "GET"
    },
    policy
  );
  assert.equal(getForm.skipReason, "get_forms_disabled");

  const allowedGetForm = classifySafeExplorerAction(
    {
      actionType: "form_get",
      label: "Search",
      text: "Search",
      selectorHint: "form#search",
      href: "",
      targetURL: "/search",
      method: "GET"
    },
    { ...policy, allowGetForms: true }
  );
  assert.equal(allowedGetForm.decision, "execute");

  const postForm = classifySafeExplorerAction(
    {
      actionType: "form_post",
      label: "Contact",
      text: "Contact",
      selectorHint: "form#contact",
      href: "",
      targetURL: "/contact",
      method: "POST"
    },
    { ...policy, allowGetForms: true }
  );
  assert.equal(postForm.skipReason, "form_method_not_safe");
});

test("duplicate actions keep a deterministic skip reason", () => {
  const safe = classifySafeExplorerAction(link("/about", "About"), policy);
  const duplicate = markSafeExplorerDuplicate(safe);
  assert.equal(duplicate.decision, "skip");
  assert.equal(duplicate.skipReason, "duplicate_url");
});

test("buildSafeExplorerActionFinding maps skip reasons to requested categories", () => {
  const external = classifySafeExplorerAction(link("https://example.com", "External"), policy);
  assert.equal(buildSafeExplorerActionFinding(external, ["evidence-id"])?.category, "explorer_external_action_skipped");

  const unsafe = classifySafeExplorerAction(link("/delete-account", "Delete"), policy);
  assert.equal(buildSafeExplorerActionFinding(unsafe, ["evidence-id"])?.category, "explorer_unsafe_action_skipped");

  const unsupported = classifySafeExplorerAction(
    { actionType: "button", label: "Open", text: "Open", selectorHint: "button", href: "", targetURL: "", method: "" },
    policy
  );
  assert.equal(buildSafeExplorerActionFinding(unsupported, ["evidence-id"])?.category, "explorer_unsupported_action");
});
