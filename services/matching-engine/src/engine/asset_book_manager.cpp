#include "asset_book_manager.hpp"

#include <stdexcept>

namespace engine {

Book& AssetBookManager::get_or_create(uint64_t asset_id) {
    if (asset_id == 0) {
        throw std::invalid_argument("asset_id 0 is a reserved sentinel, not a valid asset");
    }

    auto it = books_.find(asset_id);
    if (it != books_.end()) {
        return *it->second;
    }

    auto [inserted, _] = books_.emplace(asset_id, std::make_unique<Book>());
    return *inserted->second;
}

bool AssetBookManager::has_book(uint64_t asset_id) const {
    return books_.find(asset_id) != books_.end();
}

size_t AssetBookManager::book_count() const { return books_.size(); }

}  // namespace engine
