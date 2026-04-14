import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import ManualEntry from "../ManualEntry";

// Minimal i18next mock so the component renders in isolation.
vi.mock("react-i18next", () => ({
  useTranslation: () => ({
    t: (k: string) => k,
  }),
}));

describe("ManualEntry", () => {
  it("calls onSubmit with a normalized 13-digit EAN", () => {
    const onSubmit = vi.fn();
    render(<ManualEntry onSubmit={onSubmit} />);
    const input = screen.getByLabelText("scanner.manualLabel") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "12345678" } });
    fireEvent.click(screen.getByText("scanner.manualSubmit"));
    expect(onSubmit).toHaveBeenCalledWith("0000012345678");
  });

  it("shows an error for invalid input", () => {
    const onSubmit = vi.fn();
    render(<ManualEntry onSubmit={onSubmit} />);
    const input = screen.getByLabelText("scanner.manualLabel") as HTMLInputElement;
    fireEvent.change(input, { target: { value: "abc" } });
    fireEvent.click(screen.getByText("scanner.manualSubmit"));
    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByRole("alert")).toBeTruthy();
  });
});
