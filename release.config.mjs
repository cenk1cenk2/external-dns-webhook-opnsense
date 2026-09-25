export default {
  extends: "@cenk1cenk2/semantic-release-config",
  plugins: [["@cenk1cenk2/semantic-release-config/presets/tag", { provider: [], commit: false, changelog: false }], "@semantic-release/github"],
};
