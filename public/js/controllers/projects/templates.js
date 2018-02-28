define(['controllers/projects/taskRunner'], function () {
	app.registerController('ProjectTemplatesCtrl', ['$scope', '$http', '$uibModal', 'Project', '$rootScope', '$window', function ($scope, $http, $modal, Project, $rootScope, $window) {
		$http.get(Project.getURL() + '/keys?type=ssh').then(function (keys) {
			$scope.sshKeys = keys.data;

			$scope.sshKeysAssoc = {};
			keys.data.forEach(function (k) {
				if (k.removed) k.name = '[removed] - ' + k.name;
				$scope.sshKeysAssoc[k.id] = k;
			});
		});
		$http.get(Project.getURL() + '/inventory').then(function (inv) {
			$scope.inventory = inv.data;

			$scope.inventoryAssoc = {};
			inv.data.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;
				$scope.inventoryAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/repositories').then(function (repos) {
			$scope.repos = repos.data;

			$scope.reposAssoc = {};
			repos.data.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;

				$scope.reposAssoc[i.id] = i;
			});
		});
		$http.get(Project.getURL() + '/environment').then(function (env) {
			$scope.environment = env.data;

			$scope.environmentAssoc = {};
			env.data.forEach(function (i) {
				if (i.removed) i.name = '[removed] - ' + i.name;

				$scope.environmentAssoc[i.id] = i;
			});
		});

		function getHiddenTemplates() {
			try {
				return JSON.parse($window.localStorage.getItem('hidden-templates') || '[]');
			} catch(e) {
				return [];
			}
		}

		function setHiddenTemplates(hiddenTemplates) {
			$window.localStorage.setItem('hidden-templates', JSON.stringify(hiddenTemplates));
		}

		$scope.hasHiddenTemplates = function() {
			return getHiddenTemplates().length > 0;
		}

		$scope.reload = function () {
			$http.get(Project.getURL() + '/templates?sort=alias&order=asc').then(function (response) {
			  var templates = response.data;
				var hiddenTemplates = getHiddenTemplates();
				for (var i = 0; i < templates.length; i++) {
					var template = templates[i];
					if (hiddenTemplates.indexOf(template.id) !== -1) {
						template.hidden = true;
					}
				}
				$scope.templates = templates;
			});
		}

		$scope.remove = function (template) {
			$http.delete(Project.getURL() + '/templates/' + template.id).then(function () {
				$scope.reload();
			}).catch(function () {
				swal('error', 'could not delete template..', 'error');
			});
		}

		$scope.add = function () {
			var scope = $rootScope.$new();
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;

			$modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			}).result.then(function (opts) {
				var tpl = opts.template;
				$http.post(Project.getURL() + '/templates', tpl).then(function () {
					$scope.reload();
				}).catch(function (response) {
					swal('error', 'could not add template:' + response.status, 'error');
				});
			});
		}

		$scope.update = function (template) {
			var scope = $rootScope.$new();
			scope.tpl = template;
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;

			var modal = $modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			});

			modal.result.then(function (opts) {
				if (opts.remove) {
					return $scope.remove(template);
				}

				var tpl = opts.template;
				$http.put(Project.getURL() + '/templates/' + template.id, tpl).then(function () {
					$scope.reload();
				}).catch(function (response) {
					swal('error', 'could not add template:' + response.status, 'error');
				});
			});

			modal.closed.then(function () {
				$scope.reload();
			});
		}

		$scope.run = function (tpl) {
			$modal.open({
				templateUrl: '/tpl/projects/createTaskModal.html',
				controller: 'CreateTaskCtrl',
				resolve: {
					Project: function () {
						return Project;
					},
					Template: function () {
						return tpl;
					}
				}
			}).result.then(function (task) {
				var scope = $rootScope.$new();
				scope.task = task;
				scope.project = Project;

				$modal.open({
					templateUrl: '/tpl/projects/taskModal.html',
					controller: 'TaskCtrl',
					scope: scope,
					size: 'lg'
				});
			})
		}

		$scope.showAll = function() {
			$scope.allShown = true;
		}

		$scope.hideHidden = function() {
			$scope.allShown = false;
		}

		$scope.hideTemplate = function(template) {
			var hiddenTemplates = getHiddenTemplates();
			if (hiddenTemplates.indexOf(template.id) === -1) {
				hiddenTemplates.push(template.id);
			}
			setHiddenTemplates(hiddenTemplates);
			template.hidden = true;
		}

		$scope.showTemplate = function(template) {
			var hiddenTemplates = getHiddenTemplates();
			var i = hiddenTemplates.indexOf(template.id);
			if (i !== -1) {
				hiddenTemplates.splice(i, 1);
			}
			setHiddenTemplates(hiddenTemplates);
			delete template.hidden;
		}

		$scope.copy = function (template) {
			var tpl = angular.copy(template);
			tpl.id = null;

			var scope = $rootScope.$new();
			scope.tpl = tpl;
			scope.keys = $scope.sshKeys;
			scope.inventory = $scope.inventory;
			scope.repositories = $scope.repos;
			scope.environment = $scope.environment;

			$modal.open({
				templateUrl: '/tpl/projects/templates/add.html',
				scope: scope
			}).result.then(function (opts) {
				var tpl = opts.template;
				$http.post(Project.getURL() + '/templates', tpl).then(function () {
					$scope.reload();
				}).catch(function (response) {
					swal('error', 'could not add template:' + response.status, 'error');
				});
			});
		}

		$scope.reload();
	}]);
});
