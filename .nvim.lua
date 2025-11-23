local M = {}

if vim.tbl_contains({ "fanboy" }, vim.uv.os_gethostname()) then
  require("ck.setup").fn.setup_callback(require("ck.plugins.neotest").name, function(c)
    -- c.adapters = {
    --   require("nvim-ginkgo"),
    -- }

    return c
  end)
end

return M
