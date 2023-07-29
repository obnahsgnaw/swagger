<script>
import ThreeRoute from "./ThreeRoute";
import ThreeTitle from "./ThreeTitle";

export default {
  name: "ThreeMenu",
  functional: true,
  props: {
    menuData: {
      type: Array,
      default: () => {
        [];
      }
    },
    collapsed: {
      type: Boolean,
      default: false
    }
  },
  render(h, context) {
    let { menuData } = context.props;
    const collapsed = context.props.collapsed;

    menuData = menuData.filter(function (item){
      return item.name !== ""
    })
    menuData = menuData.map(function (item){
      if (item.sortNum !== undefined){
        return item
      }
      if (item.name.indexOf(".") !== -1){
        const sements = item.name.split(".")
        item.name = sements[1]
        item.sortNum = parseInt(sements[0])+4
      }else{
        if (item.name === "主页" || item.name === "Authorize" || item.name === "Swagger Models" || item.name === "文档管理" || item.name === "默认" || item.name === "选项管理"){
          item.sortNum = 0
        }else{
          item.sortNum = 999
        }
      }
      return item
    })
    menuData = menuData.sort(function (a,b){
      return a.sortNum-b.sortNum
    })

    const getSubMenuOrItem = (item) => {
      if (item.children && item.children.some((child) => child.name)) {
        const childrenItems = getNavMenuItems(item.children); // eslint-disable-line

        if (childrenItems && childrenItems.length > 0) {
          return (
            <a-sub-menu
              key={item.key}
              title={<ThreeTitle collapsed={collapsed} item={item} />}
            >
              {childrenItems}
            </a-sub-menu>
          );
        }
        // 当无子菜单时就不展示菜单
        return null;
      } else {
        return (
          <a-menu-item key={item.key}>
            <ThreeRoute item={item} />
          </a-menu-item>
        );
      }
    };

    const getNavMenuItems = (data) => {
      if (!data) {
        return [];
      }
      return data.map((item) => {
        return getSubMenuOrItem(item);
      });
    };
    return getNavMenuItems(menuData);
  },
};
</script>
