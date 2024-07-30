# Setting up

-   Make sure to have npm installed on your machine.
-   To install all dependencies, run `npm i` (`i` is short for `install`).
-   To run the app, run `npm start`.
-   To install new dependencies, run `npm i package-name --save`. The `--save` flag will save this dependency into package.json and package-lock.json so other developers can just run `npm i` to update their dependencies.
-   To format, run `npm run format`.
-   Recommended extension: vscode-styled-components

# Early development

To test the look and feel of your components, feel free to add your component to App.tsx. We will not style App.tsx until close to the end of the development cycle.

# Staging environment password

Tahu&Tempe5bgks

# Environment variables in deployment

To set environment variables for deployment, we add them directly to the deployment yaml file. For example, check out this file:
.github/workflows/azure-static-web-apps-production.yml
More about using environment variables in github actions here:
https://docs.github.com/en/actions/learn-github-actions/variables

# Deployment

The staging app gets deployed whenever the main branch is updated; e.g. when you push code to the main branch or when a pull request gets merged into main.
The production app gets deployed whenever the prod branch is updated; e.g. when you push code to the prod branch or when a pull request gets merged into prod.

## TL;DR

Staging branch: main
Production branch: prod

## Deployment portals

Production: https://portal.azure.com/#@tharunthirdai.onmicrosoft.com/resource/subscriptions/e16554b2-c9b5-45ee-9b36-f7f713f5f830/resourceGroups/RiceLLM/providers/Microsoft.Web/staticSites/Rice-LLM-Client/staticsite

Staging: https://portal.azure.com/#@tharunthirdai.onmicrosoft.com/resource/subscriptions/e16554b2-c9b5-45ee-9b36-f7f713f5f830/resourceGroups/RiceLLM/providers/Microsoft.Web/staticSites/Rice-LLM-Client-Staging/staticsite
