import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import "../../i18n/config";
import SearchBox from "../SearchBox";

describe("SearchBox", () => {
  it("renders a search landmark with a labelled input", () => {
    render(<SearchBox onSubmit={() => {}} />);
    expect(screen.getByRole("search")).toBeInTheDocument();
    expect(screen.getByRole("searchbox")).toBeInTheDocument();
  });

  it("calls onSubmit with trimmed value", async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(<SearchBox onSubmit={onSubmit} />);
    await user.type(screen.getByRole("searchbox"), "  mela  ");
    await user.click(screen.getByRole("button"));
    expect(onSubmit).toHaveBeenCalledWith("mela");
  });

  it("debounces onQueryChange", async () => {
    vi.useFakeTimers();
    const onChange = vi.fn();
    render(<SearchBox onSubmit={() => {}} onQueryChange={onChange} debounceMs={200} />);
    const input = screen.getByRole("searchbox");
    // userEvent with fake timers requires `advanceTimers`.
    const user = userEvent.setup({ advanceTimers: vi.advanceTimersByTime });
    await user.type(input, "pa");
    // Before debounce elapses: no call (aside from the initial empty)
    onChange.mockClear();
    vi.advanceTimersByTime(199);
    expect(onChange).not.toHaveBeenCalled();
    vi.advanceTimersByTime(2);
    expect(onChange).toHaveBeenCalledWith("pa");
    vi.useRealTimers();
  });
});
