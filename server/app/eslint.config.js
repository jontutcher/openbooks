import js from "@eslint/js";
import jsxA11y from "eslint-plugin-jsx-a11y";
import prettierRecommended from "eslint-plugin-prettier/recommended";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import globals from "globals";
import tseslint from "typescript-eslint";

export default tseslint.config(
  // Flat config no longer reads .eslintignore, so the ignores live here.
  { ignores: ["dist/**", ".vscode/**", ".idea/**", ".github/**"] },
  js.configs.recommended,
  tseslint.configs.recommended,
  react.configs.flat.recommended,
  react.configs.flat["jsx-runtime"],
  reactHooks.configs.flat.recommended,
  jsxA11y.flatConfigs.recommended,
  // Must stay last: it turns off the rules that conflict with Prettier.
  prettierRecommended,
  {
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "module",
      globals: { ...globals.browser, ...globals.node },
      parserOptions: { ecmaFeatures: { jsx: true } }
    },
    settings: { react: { version: "detect" } },
    rules: {
      // TypeScript checks the props, so the runtime propTypes are redundant.
      "react/prop-types": "off",
      // react-hooks v7 reports where the React Compiler would bail out of
      // memoizing a component. The compiler is not enabled in this build, so
      // the reports describe something that cannot happen here.
      "react-hooks/incompatible-library": "off"
    }
  }
);
