import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { SectionState } from "../components/SectionState";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";
import { formatDateTime, formatMoney } from "../lib/utils";
import type { ItemList, PickupPoint, Rental } from "../types/api";

export function DashboardPage() {
  const { user } = useAuth();
  const [items, setItems] = useState<ItemList[]>([]);
  const [rentals, setRentals] = useState<Rental[]>([]);
  const [lastPickupPoint, setLastPickupPoint] = useState<PickupPoint | null>(null);
  const [pickupForm, setPickupForm] = useState({
    address: "",
    city: "",
    is_active: true,
  });
  const [error, setError] = useState("");

  async function loadDashboard() {
    try {
      const [nextItems, nextRentals] = await Promise.all([api.getMyItems(), api.getMyRentals()]);
      setItems(nextItems);
      setRentals(nextRentals);
      setError("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Failed to load dashboard.");
    }
  }

  useEffect(() => {
    void loadDashboard();
  }, []);

  async function handleDeleteItem(itemId: string) {
    await api.deleteItem(itemId);
    await loadDashboard();
  }

  async function handleRentalStatus(rentalId: string, action: "cancel" | "pay") {
    if (action === "cancel") {
      await api.cancelRental(rentalId);
    } else {
      await api.payRental(rentalId);
    }
    await loadDashboard();
  }

  async function handleCreatePickupPoint(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const pickupPoint = await api.createPickupPoint(pickupForm);
    setLastPickupPoint(pickupPoint);
    setPickupForm({ address: "", city: "", is_active: true });
  }

  return (
    <div className="stack-page">
      <section className="panel-shell">
        <div className="section-heading">
          <div>
            <p className="eyebrow">Dashboard</p>
            <h1>Your listings and rentals</h1>
          </div>
          <Link className="button" to="/items/new">
            New item
          </Link>
        </div>
        {error ? <p className="form-error">{error}</p> : null}

        <div className="dashboard-grid">
          <section className="panel-subsection">
            <div className="section-heading tight">
              <div>
                <p className="eyebrow">My items</p>
                <h2>{items.length} listings</h2>
              </div>
            </div>
            {items.length === 0 ? <SectionState title="No listings yet" description="Create your first rentable item." /> : null}
            <div className="dashboard-list">
              {items.map((item) => (
                <article className="dashboard-card" key={item.id}>
                  <div>
                    <strong>{item.title}</strong>
                    <p>{item.category}</p>
                    <small>
                      {item.pickup_city} • {formatMoney(item.price_per_hour)} / hour
                    </small>
                  </div>
                  <div className="inline-actions">
                    <Link className="button secondary" to={`/items/${item.id}`}>
                      Open
                    </Link>
                    <Link className="button secondary" to={`/items/${item.id}/edit`}>
                      Edit
                    </Link>
                    <button className="button tertiary" onClick={() => void handleDeleteItem(item.id)} type="button">
                      Delete
                    </button>
                  </div>
                </article>
              ))}
            </div>
          </section>

          <section className="panel-subsection">
            <div className="section-heading tight">
              <div>
                <p className="eyebrow">My rentals</p>
                <h2>{rentals.length} bookings</h2>
              </div>
            </div>
            {rentals.length === 0 ? <SectionState title="No rentals yet" description="Rent an item to start building history." /> : null}
            <div className="dashboard-list">
              {rentals.map((rental) => (
                <article className="dashboard-card" key={rental.id}>
                  <div>
                    <strong>{rental.item.title}</strong>
                    <p>{rental.item.category}</p>
                    <small>
                      {formatDateTime(rental.start_at)} to {formatDateTime(rental.end_at)}
                    </small>
                    <small>
                      {rental.pickup_point.city} • {rental.pickup_point.address} • {formatMoney(rental.total_price)}
                    </small>
                  </div>
                  <div className="inline-actions">
                    <span className="badge">{rental.status}</span>
                    <button className="button secondary" onClick={() => void handleRentalStatus(rental.id, "pay")} type="button">
                      Mark paid
                    </button>
                    <button className="button tertiary" onClick={() => void handleRentalStatus(rental.id, "cancel")} type="button">
                      Cancel
                    </button>
                  </div>
                </article>
              ))}
            </div>
          </section>
        </div>
      </section>

      {user?.role === "ADMIN" ? (
        <section className="panel-shell">
          <div className="section-heading">
            <div>
              <p className="eyebrow">Admin tools</p>
              <h2>Create pickup points</h2>
            </div>
          </div>
          <form className="editor-layout compact" onSubmit={handleCreatePickupPoint}>
            <label>
              <span>Address</span>
              <input
                value={pickupForm.address}
                onChange={(event) => setPickupForm((current) => ({ ...current, address: event.target.value }))}
                required
              />
            </label>
            <label>
              <span>City</span>
              <input
                value={pickupForm.city}
                onChange={(event) => setPickupForm((current) => ({ ...current, city: event.target.value }))}
                required
              />
            </label>
            <label className="checkbox-row">
              <input
                checked={pickupForm.is_active}
                onChange={(event) => setPickupForm((current) => ({ ...current, is_active: event.target.checked }))}
                type="checkbox"
              />
              <span>Pickup point is active</span>
            </label>
            <button className="button" type="submit">
              Create pickup point
            </button>
          </form>
          {lastPickupPoint ? (
            <div className="note-panel">
              <strong>Pickup point created</strong>
              <p>Use the newly created pickup point when creating your next listing.</p>
            </div>
          ) : null}
        </section>
      ) : null}
    </div>
  );
}
