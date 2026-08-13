import { useGetRuns } from "../../openapi/generated/taskmasterComponents";
import Error from "../Error";
import Loading from "../Loading";

function RunListing() {
  const { data, isLoading, error } = useGetRuns({});

  if (error) {
    return <Error error={error} />;
  }

  if (isLoading) {
    return <Loading />;
  }

  if (!data) {
    return null;
  }

  return (
    <div>
      <h1>Runs</h1>
      <table className={"table table-striped"}>
        <thead>
          <tr>
            <th>Run #</th>
            <th>Task Name</th>
            <th>Result</th>
            <th>Start</th>
            <th>Duration</th>
          </tr>
        </thead>
        <tbody>
          {data.runs.map((run, idx) => {
            return (
              <tr key={idx}>
                <td>#{run.runNumber}</td>
                <td>{run.taskName}</td>
                <td>{run.exitCode === 0 ? "success" : "failed"}</td>
                <td>
                  {run.startTimestamp
                    ? new Date(run.startTimestamp).toLocaleString()
                    : "unknown"}
                </td>
                <td>
                  {run.endTimestamp && run.startTimestamp
                    ? `${((run.endTimestamp - run.startTimestamp) / 1000).toFixed(2)}s`
                    : "In Progress"}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

export default RunListing;
