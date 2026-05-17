import { useEffect, useMemo, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { ReviewEditor } from "../components/ReviewEditor";
import { SectionState } from "../components/SectionState";
import { useAuth } from "../context/AuthContext";
import { api } from "../lib/api";
import { formatDateTime, formatMoney, mergeItemSummary } from "../lib/utils";
import type { ItemDetails, ItemList, Review } from "../types/api";

export function ItemDetailsPage() {
  const { itemId = "" } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const [item, setItem] = useState<ItemDetails | null>(null);
  const [reviews, setReviews] = useState<Review[]>([]);
  const [renting, setRenting] = useState(false);
  const [startAt, setStartAt] = useState("");
  const [endAt, setEndAt] = useState("");
  const [error, setError] = useState("");
  const [editingReviewId, setEditingReviewId] = useState<string | null>(null);

  async function loadPage() {
    try {
      const [items, details, nextReviews] = await Promise.all([
        api.getItems(),
        api.getItem(itemId),
        api.getReviews(itemId).catch(() => []),
      ]);

      const matchedSummary = items.find((candidate) => candidate.id === itemId) as ItemList | undefined;
      setItem(mergeItemSummary(details, matchedSummary));
      setReviews(nextReviews);
      setError("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Failed to load item.");
    }
  }

  useEffect(() => {
    void loadPage();
  }, [itemId]);

  const pickupPoint = item?.pickup_point ?? item?.PickupPoint;
  const owner = item?.owner ?? item?.Owner;
  const canManageListing = Boolean(user && owner?.id && user.id === owner.id);
  const reviewCountLabel = useMemo(() => `${reviews.length} review${reviews.length === 1 ? "" : "s"}`, [reviews.length]);

  async function handleCreateRental(event: React.FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setRenting(true);
    setError("");

    try {
      await api.createRental({
        item_id: itemId,
        start_at: startAt,
        end_at: endAt,
      });
      setStartAt("");
      setEndAt("");
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Failed to create rental.");
    } finally {
      setRenting(false);
    }
  }

  async function handleCreateReview(payload: { rating: number; comment: string }) {
    await api.createReview(itemId, payload);
    setReviews(await api.getReviews(itemId));
  }

  async function handleUpdateReview(reviewId: string, payload: { rating: number; comment: string }) {
    await api.updateReview(itemId, reviewId, payload);
    setEditingReviewId(null);
    setReviews(await api.getReviews(itemId));
  }

  async function handleDeleteReview(reviewId: string) {
    await api.deleteReview(itemId, reviewId);
    setReviews(await api.getReviews(itemId));
  }

  if (error && !item) {
    return <SectionState title="Item unavailable" description={error} />;
  }

  if (!item) {
    return <SectionState title="Loading item" description="Fetching listing details and reviews." />;
  }

  return (
    <div className="stack-page">
      <section className="detail-layout">
        <article className="panel-shell spotlight">
          <div className="item-header">
            <div>
              <p className="eyebrow">{item.pickup_city || "Pickup city unavailable"}</p>
              <h1>{item.title}</h1>
              <p className="lead compact">{item.description || "No description was provided for this listing."}</p>
            </div>
            <div className="inline-actions">
              {item.is_platform_item ? <span className="badge">Platform listing</span> : null}
              {canManageListing ? (
                <Link className="button secondary" to={`/items/${itemId}/edit`}>
                  Edit item
                </Link>
              ) : null}
            </div>
          </div>

          <div className="detail-grid">
            <div className="detail-block">
              <span>Category</span>
              <strong>{item.category || "Uncategorized"}</strong>
            </div>
            <div className="detail-block">
              <span>Hourly rate</span>
              <strong>{formatMoney(item.price_per_hour)}</strong>
            </div>
            <div className="detail-block">
              <span>Daily rate</span>
              <strong>{formatMoney(item.price_per_day)}</strong>
            </div>
            <div className="detail-block">
              <span>Pickup address</span>
              <strong>{pickupPoint?.address || "Address available after selection"}</strong>
            </div>
          </div>
        </article>

        <aside className="stack-page">
          <section className="panel-shell">
            <div className="section-heading tight">
              <div>
                <p className="eyebrow">Rent this item</p>
                <h2>Schedule a rental</h2>
              </div>
            </div>
            <form className="stacked-form" onSubmit={handleCreateRental}>
              <label>
                <span>Start time</span>
                <input type="datetime-local" value={startAt} onChange={(event) => setStartAt(event.target.value)} required />
              </label>
              <label>
                <span>End time</span>
                <input type="datetime-local" value={endAt} onChange={(event) => setEndAt(event.target.value)} required />
              </label>
              <button className="button" disabled={renting || !user} type="submit">
                {user ? (renting ? "Creating rental..." : "Rent item") : "Login to rent"}
              </button>
              {user && owner?.id && user.id !== owner.id ? (
                <button
                  className="button secondary"
                  type="button"
                  onClick={() => navigate(`/chat/${owner.id}`)}
                >
                  Message owner
                </button>
              ) : null}
              {!user ? <p className="helper-text">Sign in to continue with your booking.</p> : null}
            </form>
          </section>

          <section className="panel-shell">
            <div className="section-heading tight">
              <div>
                <p className="eyebrow">Reviews</p>
                <h2>{reviewCountLabel}</h2>
              </div>
            </div>
            {user ? <ReviewEditor onSubmit={handleCreateReview} /> : <p className="helper-text">Login to post a review.</p>}
            <div className="review-list">
              {reviews.length === 0 ? <SectionState title="No reviews yet" description="Be the first to leave feedback." /> : null}
              {reviews.map((review) => {
                const isOwner = user?.id === review.reviewed_user_id;
                const isEditing = editingReviewId === review.id;

                return (
                  <article className="review-card" key={review.id}>
                    {isEditing ? (
                      <ReviewEditor
                        review={review}
                        onCancel={() => setEditingReviewId(null)}
                        onSubmit={(payload) => handleUpdateReview(review.id, payload)}
                      />
                    ) : (
                      <>
                        <div className="review-card__header">
                          <strong>{review.rating}/5</strong>
                          <span>{formatDateTime(review.created_at)}</span>
                        </div>
                        <p>{review.comment || "No comment provided."}</p>
                        {isOwner ? (
                          <div className="inline-actions">
                            <button className="button secondary" onClick={() => setEditingReviewId(review.id)} type="button">
                              Edit
                            </button>
                            <button className="button tertiary" onClick={() => void handleDeleteReview(review.id)} type="button">
                              Delete
                            </button>
                          </div>
                        ) : null}
                      </>
                    )}
                  </article>
                );
              })}
            </div>
          </section>
        </aside>
      </section>
      {error ? <p className="form-error">{error}</p> : null}
    </div>
  );
}
