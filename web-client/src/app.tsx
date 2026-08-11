import "./app.css";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { ErrorBoundary, LocationProvider, Route, Router, useLocation } from "preact-iso";
import { useEffect } from "preact/hooks";
import TaskListing from "./ui/listing/TaskListing";
import RunListing from "./ui/listing/RunListing";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    }
  }
});

const NotFound = () => (
  <div>
    <p>Page not found</p>
    <p>
      <a href="/">Home</a>
    </p>
  </div>
);

export function App() {


  return (
    <QueryClientProvider client={queryClient}>
      <LocationProvider>
        <ErrorBoundary>
          <LoadedApp />
        </ErrorBoundary>
      </LocationProvider>
    </QueryClientProvider>
  );
}

function LoadedApp() {


  return (
      <div className="container">
        <Router>
          <Route path="/tasks" component={TaskListing} />
          <Route path="/runs" component={RunListing} />
          <Route path="/" component={() => <Redirect to={"/runs"} />} />
          <NotFound default />
        </Router>
      </div>
  );
}

function Redirect({to}: {to: string}) {
  const {route} = useLocation();

  useEffect(() => route(to));
  return null;
}

