import { Middleware } from "@reduxjs/toolkit";
import { api } from "./api";
import { searchTimedOut } from "./historySlice";
import type { RootState } from "./store";

// How long to wait for results before giving up on a search. The server bounds
// its own side of the transfer well inside this, so reaching the deadline means
// the answer is not coming rather than that it is slow.
export const SEARCH_DEADLINE_MS = 120_000;

// Results are delivered over the event stream, which drops anything published
// while the browser is reconnecting, and cannot deliver at all once the server
// process the search was sent to has gone. Neither shows up as an error, so a
// search needs a deadline of its own: without one the UI shows a spinner for
// results that will never arrive, and since history is persisted, reloading
// brings that spinner straight back.
export const searchTimeoutMiddleware: Middleware =
  (store) => (next) => (action) => {
    const result = next(action);

    if (api.endpoints.search.matchPending(action)) {
      // The reducer for this action has just run, creating the history item
      // and making it the active one.
      const timestamp = (store.getState() as RootState).history.active;

      if (timestamp !== undefined) {
        setTimeout(
          () => store.dispatch(searchTimedOut(timestamp)),
          SEARCH_DEADLINE_MS
        );
      }
    }

    return result;
  };
