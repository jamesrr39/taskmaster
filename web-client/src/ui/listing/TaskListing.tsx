
import { useGetTasks } from "../../openapi/generated/taskmasterComponents";
import Error from "../Error";
import Loading from "../Loading";

function TaskListing() {
    const { data, isLoading, error } = useGetTasks({})
    if (error) {
        return <Error error={error} />
    }

    if (isLoading) {
        return <Loading />
    }

    if (!data) {
        return null;
    }

    return (
        <div>
            <h1>Tasks</h1>
            <table className={"table table-striped"}>
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Compression</th>
                        <th>Script</th>
                    </tr>
                </thead>
                <tbody>
                {data.tasks.map((task, idx) => {
                    return (
                        <tr key={idx}>
                            <td>{task.name}</td>
                            <td>{task.log.compression}</td>
                            <td><pre style={{alignItems: 'left'}}>{task.script}</pre></td>
                        </tr>
                    )
                })}
                </tbody>
            </table>
        </div>
    );
}

export default TaskListing;