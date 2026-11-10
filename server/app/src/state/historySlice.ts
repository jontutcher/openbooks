import { createSelector, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { api } from "./api";
import { BookDetail, ParseError, SearchResponse } from "./messages";
import { RootState } from "./store";

// HistoryItem represents a single search history item
export type HistoryItem = {
  query: string;
  timestamp: number;
  results?: BookDetail[];
  errors?: ParseError[];
  // Set once results can no longer arrive: the search outlived its deadline,
  // or it was still waiting when the page was loaded. Without it a search
  // that was never answered cannot be told from one still in flight, and the
  // UI waits on it forever.
  failed?: boolean;
};

export interface HistoryState {
  active: number | undefined;
  items: HistoryItem[];
}

// isPending reports whether a search is still expected to deliver results.
// Results and errors arrive together, so a finished search has both fields
// set even when either list is empty; testing the fields rather than their
// lengths keeps a search that matched nothing from looking unfinished.
export const isPending = (item: HistoryItem): boolean =>
  !item.failed && item.results === undefined && item.errors === undefined;

// The bot replaces the spaces in a query with underscores when it names the
// results file, so neither separator can be taken literally when matching the
// query a result reports against the one that was searched for.
const normalizeQuery = (query: string): string =>
  query
    .trim()
    .toLowerCase()
    .split(/[\s_]+/)
    .join(" ");

// findTarget picks the search a set of results belongs to. Several searches
// can be in flight, results can arrive in any order, and a result can name a
// search this page load has already given up on.
const findTarget = (
  items: HistoryItem[],
  query: string | undefined
): HistoryItem | undefined => {
  const wanted = normalizeQuery(query ?? "");

  if (wanted !== "") {
    const waiting = items.find(
      (item) => isPending(item) && normalizeQuery(item.query) === wanted
    );
    if (waiting !== undefined) return waiting;

    // Late: the deadline or a page load already failed it. Still better to
    // show the results under the query that asked for them.
    const named = items.find((item) => normalizeQuery(item.query) === wanted);
    if (named !== undefined) return named;
  }

  // A server too old to report the query leaves only ordering to go on. One
  // IRC connection answers searches in turn, so the oldest still waiting is
  // the one being answered. Items are newest first.
  const pending = items.filter(isPending);
  return pending[pending.length - 1];
};

const loadState = (): HistoryItem[] => {
  try {
    const items: HistoryItem[] =
      JSON.parse(localStorage.getItem("history")!) ?? [];

    // Results come over an event stream that does not replay what was missed,
    // and this load may not even be talking to the server process the search
    // was sent to, so anything still waiting can never be answered now.
    return items.map((item) =>
      isPending(item) ? { ...item, failed: true } : item
    );
  } catch {
    return [];
  }
};

const loadActiveState = (): number | undefined => {
  try {
    return JSON.parse(localStorage.getItem("active")!);
  } catch {
    return undefined;
  }
};

const initialState: HistoryState = {
  active: loadActiveState(),
  items: loadState()
};

export const historySlice = createSlice({
  name: "history",
  initialState,
  reducers: {
    setActiveItem(state, action: PayloadAction<HistoryItem | null>) {
      state.active = action.payload?.timestamp;
    },
    removeResults: (state, action: PayloadAction<number>) => {
      const timestamp = action.payload;

      if (state.active === timestamp) {
        state.active = undefined;
      }

      state.items = state.items.filter((x) => x.timestamp !== timestamp);
    },
    // searchTimedOut gives up on a search that has passed its deadline. It
    // leaves a search that has since been answered alone.
    searchTimedOut: (state, action: PayloadAction<number>) => {
      const item = state.items.find((x) => x.timestamp === action.payload);

      if (item !== undefined && isPending(item)) {
        item.failed = true;
      }
    },
    resultsReceived: (state, action: PayloadAction<SearchResponse>) => {
      const target = findTarget(state.items, action.payload.query);

      // Nowhere to put them: the search was deleted while it was in flight.
      // Dropping them beats attaching them to an unrelated query.
      if (target === undefined) return;

      target.results = action.payload.books;
      target.errors = action.payload.errors;
      delete target.failed;
    }
  },
  extraReducers: (builder) => {
    // Optimistic updates for search requests as we might get back
    // the search results before the HTTP request "completes".
    builder.addMatcher(api.endpoints.search.matchPending, (state, action) => {
      const timestamp = new Date().getTime();
      const query = action.meta.arg.originalArgs;

      state.active = timestamp;
      state.items = [{ query, timestamp }, ...state.items].slice(0, 16);
    });
    // If the request actually fails, remove the optimistic update. Match it by
    // query rather than taking the newest item, which need not be this search.
    builder.addMatcher(api.endpoints.search.matchRejected, (state, action) => {
      const query = action.meta.arg.originalArgs;
      const index = state.items.findIndex(
        (x) => isPending(x) && x.query === query
      );

      if (index === -1) return;

      if (state.active === state.items[index].timestamp) {
        state.active = undefined;
      }

      state.items.splice(index, 1);
    });
  }
});

export const selectActiveItem = createSelector(
  (state: RootState) => state.history.items,
  (state: RootState) => state.history.active,
  (items, active) => items.find((x) => x.timestamp === active)
);

export const { setActiveItem, removeResults, resultsReceived, searchTimedOut } =
  historySlice.actions;

export default historySlice.reducer;
