import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import App from "../src/App";

// Minimal smoke — fuller coverage added in later validation phase.
describe("App", () => {
  it("renders the Pesa heading", () => {
    render(<App />);
    expect(screen.getByRole("heading", { name: "Pesa" })).toBeTruthy();
  });

  it("mentions skeleton version", () => {
    render(<App />);
    expect(screen.getByText(/skeleton v0\.0\.1/i)).toBeTruthy();
  });
});
