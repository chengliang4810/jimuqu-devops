import nextPlugin from "@next/eslint-plugin-next";
import tseslint from "@typescript-eslint/eslint-plugin";
import importPlugin from "eslint-plugin-import";
import jsxA11yPlugin from "eslint-plugin-jsx-a11y";
import reactPlugin from "eslint-plugin-react";
import reactHooksPlugin from "eslint-plugin-react-hooks";

const eslintConfig = [
  {
    ignores: [".next/**", ".next-dev/**", "out/**", "node_modules/**"],
  },
  tseslint.configs["flat/base"],
  reactPlugin.configs.flat.recommended,
  reactPlugin.configs.flat["jsx-runtime"],
  reactHooksPlugin.configs["recommended-latest"],
  nextPlugin.flatConfig.coreWebVitals,
  {
    files: ["**/*.{js,jsx,ts,tsx}"],
    plugins: {
      import: importPlugin,
      "jsx-a11y": jsxA11yPlugin,
    },
    settings: {
      react: {
        version: "detect",
      },
      "import/resolver": {
        node: true,
        typescript: true,
      },
    },
    rules: {
      "import/no-anonymous-default-export": "warn",
      "react/no-unknown-property": "off",
      "react/prop-types": "off",
      "react/jsx-no-target-blank": "off",
      "jsx-a11y/alt-text": [
        "warn",
        {
          elements: ["img"],
          img: ["Image"],
        },
      ],
      "jsx-a11y/aria-props": "warn",
      "jsx-a11y/aria-proptypes": "warn",
      "jsx-a11y/aria-unsupported-elements": "warn",
      "jsx-a11y/role-has-required-aria-props": "warn",
      "jsx-a11y/role-supports-aria-props": "warn",
    },
  },
];

export default eslintConfig;
