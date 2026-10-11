import { fireEvent, render, screen } from "@testing-library/react";
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

  it("debounces onQueryChange", () => {
    vi.useFakeTimers();
    try {
      const onChange = vi.fn();
      render(<SearchBox onSubmit={() => {}} onQueryChange={onChange} debounceMs={200} />);
      // fireEvent, not userEvent: userEvent's internal delays stall under fake timers.
      fireEvent.change(screen.getByRole("searchbox"), { target: { value: "pa" } });
      onChange.mockClear(); // drop the initial empty-query call
      vi.advanceTimersByTime(199);
      expect(onChange).not.toHaveBeenCalled();
      vi.advanceTimersByTime(2);
      expect(onChange).toHaveBeenCalledWith("pa");
    } finally {
      vi.useRealTimers();
    }
  });
});
