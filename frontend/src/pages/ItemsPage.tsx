import { useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { ItemCard } from "../components/ItemCard";
import { SectionState } from "../components/SectionState";
import { api } from "../lib/api";
import type { ItemList } from "../types/api";

export function ItemsPage() {
  const [items, setItems] = useState<ItemList[]>([]);
  const [search, setSearch] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    setLoading(true);
    void api
      .getItems()
      .then((response) => {
        setItems(response);
        setError("");
      })
      .catch((requestError: Error) => setError(requestError.message))
      .finally(() => setLoading(false));
  }, []);

  const filteredItems = useMemo(() => {
    const query = search.trim().toLowerCase();
    if (!query) {
      return items;
    }

    return items.filter(
      (item) =>
        item.title.toLowerCase().includes(query) ||
        item.category.toLowerCase().includes(query) ||
        item.pickup_city.toLowerCase().includes(query),
    );
  }, [items, search]);

  return (
    <div className="stack-page">
      <section className="panel-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Marketplace</p>
            <h1>Browse rentable items</h1>
          </div>
          <Link className="button secondary" to="/items/new">
            Create listing
          </Link>
        </div>
        <div className="toolbar">
          <input
            className="search-input"
            placeholder="Search by title, category, or city"
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
        </div>
        {loading ? <SectionState title="Loading items" description="Fetching the live catalog." /> : null}
        {error ? <SectionState title="Could not load items" description={error} /> : null}
        {!loading && !error && filteredItems.length === 0 ? (
          <SectionState title="No matches" description="Try a different search term." />
        ) : null}
        {!loading && !error ? (
          <div className="card-grid">
            {filteredItems.map((item) => (
              <ItemCard key={item.id} item={item} />
            ))}
          </div>
        ) : null}
      </section>
    </div>
  );
}
