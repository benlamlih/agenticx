import js from "@eslint/js";
import globals from "globals";
import tseslint from "typescript-eslint";
import vue from "eslint-plugin-vue";
import prettier from "eslint-config-prettier";
import {defineConfig} from "eslint/config";

export default defineConfig([
    js.configs.recommended,
    {
        files: ["**/*.{js,ts,vue}"],
        languageOptions: {
            globals: globals.browser,
            parser: tseslint.parser,
            parserOptions: {
                ecmaVersion: "latest",
                sourceType: "module",
                project: "./tsconfig.json",
                extraFileExtensions: [".vue"],
            },
        },
        plugins: {
            "@typescript-eslint": tseslint.plugin,
            vue,
        },
        rules: {
            ...tseslint.configs.recommendedTypeChecked.rules,
        },
    },

    vue.configs["flat/recommended"],
    prettier,
]);
