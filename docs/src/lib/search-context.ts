import { createContext, useContext } from "react";

export const SearchContext = createContext<{ openSearch: () => void }>({ openSearch: () => {} });
export function useSearch() { return useContext(SearchContext); }
