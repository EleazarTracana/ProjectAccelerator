import { type RouteConfig, route, layout } from "@react-router/dev/routes"

export default [
  layout("./pages/app/routes/_app-layout.tsx", [
    route("/", "./pages/app/routes/_index.tsx", { id: "dashboard" }),
    layout("./pages/app/routes/reports/_layout.tsx", [
      route("/reports", "./pages/app/routes/reports/_index.tsx", {
        id: "reports",
      }),
      route("/reports/activity", "./pages/app/routes/reports/activity.tsx", {
        id: "reports-activity",
      }),
    ]),
  ]),
  route("*", "./pages/app/routes/$.tsx", { id: "not-found" }),
] satisfies RouteConfig
