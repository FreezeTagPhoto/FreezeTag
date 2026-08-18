# FreezeTag Frontend

This is a [Next.js](https://nextjs.org) project.

## Getting Started

First, run the development server:

```bash
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

Instructions for running this project alongside the backend are TBD.

This project is structured such that pages are handled in `src/app` using Next's regular folder routing system. Components and such that aren't related to a specific page should be handled in subfolders of `src` that are NOT `app`. `src/app` should be as specifically related to pages as is possible.

## Contributing

Read our project's [Contribution Guidelines](https://capstone.cs.utah.edu/groups/freezetag/-/wikis/Contribution-Guidelines).

For the frontend specifically, run these tools before opening or modifying an MR (in no particular order):

```bash
pnpm lint
pnpm format:check
pnpm test
```

All of these should pass without errors. For now, `pnpm lint` warnings are not blocking on the chain, but you should avoid submitting code that has lint warnings just-because. If you must have a lint warning, mention it in your MR. We are trying to assess the applicability of rules as they come up. Ideally, we will reach a state where the pointless/annoying/etc warnings are removed entirely, and the helpful/important warnings are proper errors.

You can run

```bash
pnpm format
```

to automatically fix formatting issues.

### BETA: Frontend API Client Codegen

We're testing and implementing Hey API's openapi-ts for generating frontend API clients based off of the backend spec, automatically. All you need to do in the frontend is use the generated files in `src/openapi-ts_beta_api` for any NEW API endpoints, or for any OLD API endpoints that have already been migrated. You are not expected to solve the migration for old, unmigrated endpoints unless you are taking on the specific issue to do the migration. If it's helpful or easy to do the migration alongside your change, assign yourself the related issue.

The default client should be OK, and you should not need any of the options for the client. I believe all you need is the exported functions in `sdk.gen.ts`.

## Next.js Resources

To learn more about Next.js, take a look at the following resources:

- [Next.js Documentation](https://nextjs.org/docs) - learn about Next.js features and API.
- [Learn Next.js](https://nextjs.org/learn) - an interactive Next.js tutorial.

You can check out [the Next.js GitHub repository](https://github.com/vercel/next.js) - your feedback and contributions are welcome!

## Testing

Simply run

```bash
pnpm test
```

This will print coverage but also put coverage into the `coverage` directory.

Tests are handled in `__tests__` and the folder structure in `__tests__` should mirror the folder structure in `src`. So, the tests for `src/app/gallery/page.tsx` should ideally be in `__tests__/app/gallery/page.tsx`. You should **always** have adequate test coverage for your PRs, as much as it is possible to do so. Variations from these rules are rarely granted and should be seldom requested.
