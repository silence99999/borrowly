export function SectionState({
  title,
  description,
}: {
  title: string;
  description: string;
}) {
  return (
    <div className="empty-card">
      <h3>{title}</h3>
      <p>{description}</p>
    </div>
  );
}
