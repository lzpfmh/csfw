// +build ignore

package cms

import (
	"github.com/corestoreio/csfw/config/model"
)

// PathWebDefaultCmsHomePage => CMS Home Page.
// SourceModel: Otnegam\Cms\Model\Config\Source\Page
var PathWebDefaultCmsHomePage = model.NewStr(`web/default/cms_home_page`, model.WithPkgCfg(PackageConfiguration))

// PathWebDefaultCmsNoRoute => CMS No Route Page.
// SourceModel: Otnegam\Cms\Model\Config\Source\Page
var PathWebDefaultCmsNoRoute = model.NewStr(`web/default/cms_no_route`, model.WithPkgCfg(PackageConfiguration))

// PathWebDefaultCmsNoCookies => CMS No Cookies Page.
// SourceModel: Otnegam\Cms\Model\Config\Source\Page
var PathWebDefaultCmsNoCookies = model.NewStr(`web/default/cms_no_cookies`, model.WithPkgCfg(PackageConfiguration))

// PathWebDefaultShowCmsBreadcrumbs => Show Breadcrumbs for CMS Pages.
// SourceModel: Otnegam\Config\Model\Config\Source\Yesno
var PathWebDefaultShowCmsBreadcrumbs = model.NewBool(`web/default/show_cms_breadcrumbs`, model.WithPkgCfg(PackageConfiguration))

// PathWebBrowserCapabilitiesCookies => Redirect to CMS-page if Cookies are Disabled.
// SourceModel: Otnegam\Config\Model\Config\Source\Yesno
var PathWebBrowserCapabilitiesCookies = model.NewBool(`web/browser_capabilities/cookies`, model.WithPkgCfg(PackageConfiguration))

// PathWebBrowserCapabilitiesJavascript => Show Notice if JavaScript is Disabled.
// SourceModel: Otnegam\Config\Model\Config\Source\Yesno
var PathWebBrowserCapabilitiesJavascript = model.NewBool(`web/browser_capabilities/javascript`, model.WithPkgCfg(PackageConfiguration))

// PathWebBrowserCapabilitiesLocalStorage => Show Notice if Local Storage is Disabled.
// SourceModel: Otnegam\Config\Model\Config\Source\Yesno
var PathWebBrowserCapabilitiesLocalStorage = model.NewBool(`web/browser_capabilities/local_storage`, model.WithPkgCfg(PackageConfiguration))

// PathCmsWysiwygEnabled => Enable WYSIWYG Editor.
// SourceModel: Otnegam\Cms\Model\Config\Source\Wysiwyg\Enabled
var PathCmsWysiwygEnabled = model.NewStr(`cms/wysiwyg/enabled`, model.WithPkgCfg(PackageConfiguration))