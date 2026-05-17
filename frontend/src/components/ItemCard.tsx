import { Link } from "react-router-dom";
import type { ItemList } from "../types/api";
import { formatMoney } from "../lib/utils";

export function ItemCard({ item }: { item: ItemList }) {
  return (
    <article className="item-card">
      <div className="item-card__meta">
        <span className="eyebrow">{item.pickup_city || "City unavailable"}</span>
        {item.is_platform_item ? <span className="badge">Platform</span> : null}
      </div>
      <h3>{item.title}</h3>
      <p>{item.category}</p>
      <dl className="price-grid">
        <div>
          <dt>Hour</dt>
          <dd>{formatMoney(item.price_per_hour)}</dd>
        </div>
        <div>
          <dt>Day</dt>
          <dd>{formatMoney(item.price_per_day)}</dd>
        </div>
      </dl>
      <Link className="button secondary" to={`/items/${item.id}`}>
        View details
      </Link>
    </article>
  );
}
