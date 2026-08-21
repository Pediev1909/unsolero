import { createBrowserRouter } from 'react-router-dom'

import { HomePage } from '../pages/HomePage'
import { ProtectedRoute } from '../features/auth/ProtectedRoute'
import { ProtectedAdminRoute } from '../features/admin/ProtectedAdminRoute'
import { RouteShell } from './RouteShell'
import { RouteLoadingState } from './RouteLoadingState'
import { RouteErrorPage } from './RouteErrorPage'

export const router = createBrowserRouter([
  {
    element: <RouteShell />,
    errorElement: <RouteErrorPage />,
    hydrateFallbackElement: <RouteLoadingState />,
    children: [
      {
        path: '/',
        element: <HomePage />,
      },
      {
        path: '/privacy',
        lazy: async () => ({
          Component: (await import('../pages/PrivacyPage')).PrivacyPage,
        }),
      },
      {
        path: '/about',
        lazy: async () => ({
          Component: (await import('../pages/AboutPage')).AboutPage,
        }),
      },
      {
        path: '/author/:slug',
        lazy: async () => ({
          Component: (await import('../pages/AuthorPage')).AuthorPage,
        }),
      },
      {
        path: '/affiliate-disclosure',
        lazy: async () => ({
          Component: (await import('../pages/AffiliateDisclosurePage'))
            .AffiliateDisclosurePage,
        }),
      },
      {
        path: '/login',
        lazy: async () => ({
          Component: (await import('../pages/LoginPage')).LoginPage,
        }),
      },
      {
        path: '/register',
        lazy: async () => ({
          Component: (await import('../pages/RegisterPage')).RegisterPage,
        }),
      },
      {
        path: '/check-email',
        lazy: async () => ({
          Component: (await import('../pages/CheckEmailPage')).CheckEmailPage,
        }),
      },
      {
        path: '/verify-email',
        lazy: async () => ({
          Component: (await import('../pages/VerifyEmailPage')).VerifyEmailPage,
        }),
      },
      {
        path: '/forgot-password',
        lazy: async () => ({
          Component: (await import('../pages/ForgotPasswordPage'))
            .ForgotPasswordPage,
        }),
      },
      {
        path: '/reset-password',
        lazy: async () => ({
          Component: (await import('../pages/ResetPasswordPage'))
            .ResetPasswordPage,
        }),
      },
      {
        path: '/login/mfa',
        lazy: async () => ({
          Component: (await import('../pages/MfaLoginPage')).MfaLoginPage,
        }),
      },
      {
        path: '/build',
        lazy: async () => ({
          Component: (await import('../pages/RecommendationBuilderPage'))
            .RecommendationBuilderPage,
        }),
      },
      {
        path: '/categories',
        lazy: async () => ({
          Component: (await import('../pages/CategoriesPage')).CategoriesPage,
        }),
      },
      {
        path: '/brands',
        lazy: async () => ({
          Component: (await import('../pages/BrandsPage')).BrandsPage,
        }),
      },
      {
        path: '/how-it-works',
        lazy: async () => ({
          Component: (await import('../pages/HowItWorksPage')).HowItWorksPage,
        }),
      },
      {
        path: '/products',
        lazy: async () => ({
          Component: (await import('../pages/ProductsPage')).ProductsPage,
        }),
      },
      {
        path: '/comparisons',
        lazy: async () => ({
          Component: (await import('../pages/ContentHubPage')).ComparisonsPage,
        }),
      },
      {
        path: '/compare',
        lazy: async () => ({
          Component: (await import('../pages/ComparePage')).ComparePage,
        }),
      },
      {
        path: '/compare/:slug',
        lazy: async () => ({
          Component: (await import('../pages/ContentDetailPage'))
            .ContentDetailPage,
        }),
      },
      {
        path: '/guides',
        lazy: async () => ({
          Component: (await import('../pages/ContentHubPage')).GuidesPage,
        }),
      },
      {
        path: '/guides/:slug',
        lazy: async () => ({
          Component: (await import('../pages/ContentDetailPage'))
            .ContentDetailPage,
        }),
      },
      {
        path: '/articles',
        lazy: async () => ({
          Component: (await import('../pages/ContentHubPage')).ArticlesPage,
        }),
      },
      {
        path: '/articles/:slug',
        lazy: async () => ({
          Component: (await import('../pages/ContentDetailPage'))
            .ContentDetailPage,
        }),
      },
      {
        path: '/wishlist',
        lazy: async () => ({
          Component: (await import('../pages/WishlistPage')).WishlistPage,
        }),
      },
      {
        path: '/setups',
        lazy: async () => ({
          Component: (await import('../pages/SavedSetupsPage')).SavedSetupsPage,
        }),
      },
      {
        path: '/setups/:setupID',
        lazy: async () => ({
          Component: (await import('../pages/SetupPage')).SetupPage,
        }),
      },
      {
        path: '/products/:slug',
        lazy: async () => ({
          Component: (await import('../pages/ProductDetailPage'))
            .ProductDetailPage,
        }),
      },
      {
        path: '/categories/:slug',
        lazy: async () => ({
          Component: (await import('../pages/CategoryPage')).CategoryPage,
        }),
      },
      {
        path: '/brands/:slug',
        lazy: async () => ({
          Component: (await import('../pages/BrandPage')).BrandPage,
        }),
      },
      {
        path: '/design-system',
        lazy: async () => ({
          Component: (await import('../pages/design-system/DesignSystemPage'))
            .DesignSystemPage,
        }),
      },
      {
        element: <ProtectedRoute />,
        children: [
          {
            path: '/account',
            lazy: async () => ({
              Component: (await import('../pages/AccountPage')).AccountPage,
            }),
          },
        ],
      },
      {
        element: <ProtectedAdminRoute />,
        children: [
          {
            path: '/admin',
            lazy: async () => ({
              Component: (
                await import('../features/admin/components/AdminLayout')
              ).AdminLayout,
            }),
            children: [
              {
                index: true,
                lazy: async () => ({
                  Component: (await import('../pages/admin/AdminDashboardPage'))
                    .AdminDashboardPage,
                }),
              },
              {
                path: 'products',
                lazy: async () => ({
                  Component: (await import('../pages/admin/AdminProductsPage'))
                    .AdminProductsPage,
                }),
              },
              {
                path: 'products/new',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminProductEditorPage')
                  ).AdminProductEditorPage,
                }),
              },
              {
                path: 'products/:productID',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminProductEditorPage')
                  ).AdminProductEditorPage,
                }),
              },
              {
                path: 'evidence',
                lazy: async () => ({
                  Component: (await import('../pages/admin/AdminEvidencePage'))
                    .AdminEvidencePage,
                }),
              },
              {
                path: 'evidence/:productID',
                lazy: async () => ({
                  Component: (await import('../pages/admin/AdminEvidencePage'))
                    .AdminEvidenceDetailPage,
                }),
              },
              {
                path: 'categories',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminReferencePages')
                  ).AdminCategoriesPage,
                }),
              },
              {
                path: 'brands',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminReferencePages')
                  ).AdminBrandsPage,
                }),
              },
              {
                path: 'merchants',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminReferencePages')
                  ).AdminMerchantsPage,
                }),
              },
              {
                path: 'offers',
                lazy: async () => ({
                  Component: (await import('../pages/admin/AdminOffersPage'))
                    .AdminOffersPage,
                }),
              },
              {
                path: 'commerce',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminCommerceOperationsPage')
                  ).AdminCommerceOperationsPage,
                }),
              },
              {
                path: 'affiliate-links',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminAffiliateLinksPage')
                  ).AdminAffiliateLinksPage,
                }),
              },
              {
                path: 'recommendations',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminRecommendationsPage')
                  ).AdminRecommendationsPage,
                }),
              },
              {
                path: 'recommendations/:recommendationID',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminRecommendationsPage')
                  ).AdminRecommendationDetailPage,
                }),
              },
              {
                path: 'users',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminReferencePages')
                  ).AdminUsersPage,
                }),
              },
              {
                path: 'events',
                lazy: async () => ({
                  Component: (await import('../pages/admin/AdminEventsPage'))
                    .AdminEventsPage,
                }),
              },
              {
                path: 'content',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminEmptySectionPage')
                  ).AdminContentPage,
                }),
              },
              {
                path: 'settings',
                lazy: async () => ({
                  Component: (
                    await import('../pages/admin/AdminEmptySectionPage')
                  ).AdminSettingsPage,
                }),
              },
            ],
          },
        ],
      },
      {
        path: '*',
        lazy: async () => ({
          Component: (await import('../pages/NotFoundPage')).NotFoundPage,
        }),
      },
    ],
  },
])
