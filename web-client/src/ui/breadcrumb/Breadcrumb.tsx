import HomeIcon from "bootstrap-icons/icons/house-door.svg";

type Props = {
  pathFragments: string[];
};

function Breadcrumb({ pathFragments }: Props) {
  const fragments = pathFragments.map((fragment, idx) => {
    return {
      title: fragment,
      url:
        "/" +
        pathFragments
          .slice(0, idx + 1)
          .map((fragment) => encodeURIComponent(fragment))
          .join("/"),
    };
  });

  return (
    <div>
      <img src={HomeIcon} />
      {fragments.map((fragment) => (
        <>
          <span> / </span>
          <a href={fragment.url}>{fragment.title}</a>
        </>
      ))}
    </div>
  );
}

export default Breadcrumb;
