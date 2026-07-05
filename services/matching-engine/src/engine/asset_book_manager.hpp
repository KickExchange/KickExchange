#ifndef ENGINE_ASSET_BOOK_MANAGER_HPP
#define ENGINE_ASSET_BOOK_MANAGER_HPP

#include "../limit_order_book/Book.hpp"

#include <cstdint>
#include <memory>
#include <unordered_map>

namespace engine {

// Routes orders to the Book instance for their asset, one Book per asset_id.
// The matching engine holds no asset metadata (name, existence) - that's
// Trading Service's job via Postgres. This just multiplexes order flow.
class AssetBookManager {
public:
    // Returns the Book for asset_id, creating one on first use.
    // Throws std::invalid_argument if asset_id == 0 (reserved/unset sentinel).
    Book& get_or_create(uint64_t asset_id);

    bool has_book(uint64_t asset_id) const;
    size_t book_count() const;

private:
    std::unordered_map<uint64_t, std::unique_ptr<Book>> books_;
};

}  // namespace engine

#endif
