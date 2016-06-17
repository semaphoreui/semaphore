app.config(function ($stateProvider, $couchPotatoProvider) {
	$stateProvider
	.state('project', {
		url: '/project/:project_id',
		abstract: true,
		templateUrl: '/tpl/projects/container.html',
		controller: function ($scope, Project) {
			$scope.project = Project;
		},
		resolve: {
			Project: ['$http', '$stateParams', '$q', 'ProjectFactory', function ($http, params, $q, ProjectFactory) {
				var d = $q.defer();

				$http.get('/project/' + params.project_id)
				.success(function (project) {
					d.resolve(new ProjectFactory(project));
				}).error(function () {
					d.resolve(false);
				});

				return d.promise;
			}]
		}
	})
	.state('project.edit', {
		url: '/edit',
		pageTitle: 'Edit Project',
		templateUrl: '/tpl/projects/edit.html',
		controller: 'ProjectEditCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/edit'])
		}
	})
	.state('project.dashboard', {
		url: '',
		pageTitle: 'Project Dashboard',
		templateUrl: '/tpl/projects/dashboard.html',
		controller: 'ProjectDashboardCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/dashboard'])
		}
	})

	.state('project.users', {
		url: '/users',
		pageTitle: 'Users',
		templateUrl: '/tpl/projects/users/list.html',
		controller: 'ProjectUsersCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/users'])
		}
	})

	.state('project.templates', {
		url: '/templates',
		pageTitle: 'Templates',
		templateUrl: '/tpl/projects/templates/list.html',
		controller: 'ProjectTemplatesCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/templates'])
		}
	})

	.state('project.inventory', {
		url: '/inventory',
		pageTitle: 'Inventory',
		templateUrl: '/tpl/projects/inventory/list.html',
		controller: 'ProjectInventoryCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/inventory'])
		}
	})

	.state('project.environment', {
		url: '/environment',
		pageTitle: 'Environment',
		templateUrl: '/tpl/projects/environment/list.html',
		controller: 'ProjectEnvironmentCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/environment'])
		}
	})

	.state('project.keys', {
		url: '/keys',
		pageTitle: 'Keys',
		templateUrl: '/tpl/projects/keys/list.html',
		controller: 'ProjectKeysCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/keys'])
		}
	})

	.state('project.repositories', {
		url: '/repositories',
		pageTitle: 'Repositories',
		templateUrl: '/tpl/projects/repositories/list.html',
		controller: 'ProjectRepositoriesCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/repositories'])
		}
	})

	.state('project.schedule', {
		url: '/schedule',
		pageTitle: 'Template Schedule',
		templateUrl: '/tpl/projects/schedule.html',
		controller: 'ProjectScheduleCtrl',
		resolve: {
			$d: $couchPotatoProvider.resolveDependencies(['controllers/projects/schedule'])
		}
	});
});