import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { ItemCard } from "../components/ItemCard";
import { api } from "../lib/api";
import type { ItemList } from "../types/api";

export function HomePage() {
  const [items, setItems] = useState<ItemList[]>([]);

  useEffect(() => {
    void api.getItems().then((data) => setItems(data.slice(0, 3))).catch(() => setItems([]));
  }, []);

  return (
    <div className="stack-page">
      <section className="hero-panel">
        <div className="hero-copy">
          <p className="eyebrow">Rent Items Platform</p>
          <h1>Turn idle gear into bookable inventory.</h1>
          <p className="lead">
            Browse public listings, manage your own rentable items, and track rentals from a single dashboard.
          </p>
          <div className="inline-actions">
            <Link className="button" to="/items">
              Explore items
            </Link>
            <Link className="button secondary" to="/items/new">
              List an item
            </Link>
          </div>
        </div>
        <aside className="hero-aside">
          <div className="metric-card">
            <strong>{items.length || 0}</strong>
            <span>Featured listings available right now</span>
          </div>
          <div className="metric-card accent">
            <strong>Fast access</strong>
            <span>Browse listings, manage rentals, and keep everything in one place.</span>
          </div>
        </aside>
      </section>

      <section className="panel-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Featured</p>
            <h2>Recently loaded inventory</h2>
          </div>
          <Link className="text-link" to="/items">
            See all items
          </Link>
        </div>
        <div className="card-grid">
          {items.map((item) => (
            <ItemCard key={item.id} item={item} />
          ))}
        </div>
      </section>
    </div>
  );
}
