(function () {
  function text(value, fallback) {
    return value || fallback || "";
  }

  function nodeDataFrom(element) {
    return {
      id: element.dataset.uweId,
      name: text(element.dataset.uweLabel, element.dataset.uweId),
      stereo: text(element.dataset.uweStereo, "«navigationClass»"),
      type: text(element.dataset.uweType, "navigation"),
      packageName: text(element.dataset.uwePackage, "Navigation"),
      route: text(element.dataset.uweRoute, ""),
      role: text(element.dataset.uweRole, ""),
      actions: text(element.dataset.uweActions, "Inspect available actions in the source matrix."),
      effect: text(element.dataset.uweEffect, "Effect not recorded."),
      screenshot: text(element.dataset.uweScreenshot, "")
    };
  }

  function edgeDataFrom(element) {
    return {
      id: element.dataset.uweFrom + "__" + element.dataset.uweTo + "__" + text(element.dataset.uweLabel, "link"),
      source: element.dataset.uweFrom,
      target: element.dataset.uweTo,
      label: text(element.dataset.uweLabel, "«navigationLink»")
    };
  }

  function classFor(node) {
    var stereo = String(node.stereo || "").toLowerCase();
    var facets = String(node.type || "").toLowerCase();
    if (stereo.indexOf("processclass") >= 0) return "uwe-process";
    if (facets.indexOf("adaptation") >= 0 || facets.indexOf("denied") >= 0) return "uwe-adaptation";
    if (facets.indexOf("access") >= 0) return "uwe-access";
    return "uwe-navigation";
  }

  function setInspector(workspace, data) {
    workspace.querySelector("[data-uwe-inspector-stereo]").textContent = data.stereo || "UWE node";
    workspace.querySelector("[data-uwe-inspector-title]").textContent = data.name || data.id;
    workspace.querySelector("[data-uwe-inspector-package]").textContent = data.packageName || "Navigation";
    workspace.querySelector("[data-uwe-inspector-route]").textContent = data.route || "state";
    workspace.querySelector("[data-uwe-inspector-role]").textContent = data.role || "all roles";
    workspace.querySelector("[data-uwe-inspector-actions]").textContent = data.actions || "No action inventory recorded.";
    workspace.querySelector("[data-uwe-inspector-effect]").textContent = data.effect || "No side effect recorded.";
    var img = workspace.querySelector("[data-uwe-inspector-image]");
    if (img && data.screenshot) {
      img.src = data.screenshot;
      img.alt = (data.name || data.id) + " screenshot";
      img.dataset.uweCaption = (data.stereo || "UWE node") + " " + (data.name || data.id) + ": " + (data.effect || "");
    }
  }

  function setActiveNodeButton(workspace, id) {
    workspace.querySelectorAll("[data-uwe-focus-node]").forEach(function (button) {
      button.classList.toggle("active", button.dataset.uweFocusNode === id);
    });
  }

  function setActivePackageButton(workspace, packageName) {
    workspace.querySelectorAll("[data-uwe-action]").forEach(function (button) {
      var action = button.dataset.uweAction || "";
      button.classList.toggle("active", action === "package:" + packageName);
    });
  }

  function openLightbox(workspace) {
    var img = workspace.querySelector("[data-uwe-inspector-image]");
    var lightbox = workspace.querySelector("[data-uwe-lightbox]");
    var lightboxImg = workspace.querySelector("[data-uwe-lightbox-image]");
    var caption = workspace.querySelector("[data-uwe-lightbox-caption]");
    if (!img || !lightbox || !lightboxImg || !img.src) return;
    lightboxImg.src = img.src;
    lightboxImg.alt = img.alt || "Selected UWE screenshot";
    if (caption) caption.textContent = img.dataset.uweCaption || img.alt || "Selected UWE screenshot";
    lightbox.classList.add("active");
  }

  function closeLightbox(workspace) {
    var lightbox = workspace.querySelector("[data-uwe-lightbox]");
    if (lightbox) lightbox.classList.remove("active");
  }

  function setFocusMode(workspace, enabled, cy) {
    workspace.classList.toggle("uwe-focus-mode", enabled);
    document.documentElement.classList.toggle("uwe-focus-active", enabled);
    var button = workspace.querySelector("[data-uwe-action=workspace-focus]");
    if (button) button.textContent = enabled ? "Exit focus" : "Focus workspace";
    setTimeout(function () {
      if (cy) {
        cy.resize();
        cy.fit(undefined, 42);
      }
    }, 80);
  }

  function initPanZoom() {
    if (!window.svgPanZoom) return;
    document.querySelectorAll("[data-svg-pan-zoom=true] svg").forEach(function (svg) {
      window.svgPanZoom(svg, {
        controlIconsEnabled: true,
        fit: true,
        center: true,
        minZoom: 0.1,
        maxZoom: 20,
        zoomScaleSensitivity: 0.25
      });
    });
    document.querySelectorAll(".viewer-badge").forEach(function (el) {
      el.textContent = "svg-pan-zoom active: drag, wheel, +/- controls";
    });
  }

  function initWorkspace(workspace) {
    var graphHost = workspace.querySelector("[data-uwe-cy]");
    var badge = workspace.querySelector("[data-uwe-runtime-badge]");
    var nodeElements = Array.prototype.slice.call(workspace.querySelectorAll("[data-uwe-node]"));
    var edgeElements = Array.prototype.slice.call(workspace.querySelectorAll("[data-uwe-edge]"));
    var nodes = nodeElements.map(nodeDataFrom);
    var edges = edgeElements.map(edgeDataFrom).filter(function (edge) {
      return nodes.some(function (node) { return node.id === edge.source; }) && nodes.some(function (node) { return node.id === edge.target; });
    });
    if (!graphHost || !window.cytoscape) {
      if (badge) badge.textContent = "Graph workspace fallback: Cytoscape runtime unavailable";
      return;
    }
    if (window.cytoscapeDagre) window.cytoscape.use(window.cytoscapeDagre);
    graphHost.innerHTML = "";
    var cy = window.cytoscape({
      container: graphHost,
      elements: nodes.map(function (node) {
        return { group: "nodes", data: node, classes: classFor(node) };
      }).concat(edges.map(function (edge) {
        return { group: "edges", data: edge };
      })),
      minZoom: 0.12,
      maxZoom: 2.8,
      wheelSensitivity: 0.18,
      style: [
        {
          selector: "node",
          style: {
            shape: "round-rectangle",
            width: 232,
            height: 168,
            "background-color": "#ffffff",
            "background-image": "data(screenshot)",
            "background-fit": "contain",
            "background-opacity": 0.92,
            "border-color": "#111827",
            "border-width": 1.5,
            label: "data(name)",
            "font-family": "Segoe UI, Arial, sans-serif",
            "font-weight": 800,
            "font-size": 12,
            color: "#111827",
            "text-background-color": "#ffffff",
            "text-background-opacity": 0.92,
            "text-background-padding": 5,
            "text-valign": "bottom",
            "text-halign": "center",
            "overlay-opacity": 0
          }
        },
        { selector: ".uwe-process", style: { "border-color": "#a15c07", "border-style": "dashed" } },
        { selector: ".uwe-access", style: { "border-color": "#2457c5" } },
        { selector: ".uwe-adaptation", style: { "border-color": "#b42318", "border-style": "dashed" } },
        {
          selector: "edge",
          style: {
            width: 2,
            "line-color": "#64748b",
            "target-arrow-color": "#64748b",
            "target-arrow-shape": "triangle",
            "curve-style": "bezier",
            label: "data(label)",
            "font-size": 10,
            "font-family": "Segoe UI, Arial, sans-serif",
            color: "#334155",
            "text-background-color": "#ffffff",
            "text-background-opacity": 0.85,
            "text-background-padding": 2
          }
        },
        { selector: ":selected", style: { "border-width": 4, "border-color": "#0f766e", "line-color": "#0f766e", "target-arrow-color": "#0f766e" } }
      ],
      layout: {
        name: window.cytoscapeDagre ? "dagre" : "grid",
        rankDir: "LR",
        nodeSep: 70,
        edgeSep: 24,
        rankSep: 132,
        fit: true,
        padding: 36
      }
    });
    workspace.uweCy = cy;
    workspace.querySelector("[data-uwe-stat=nodes]").textContent = String(nodes.length);
    workspace.querySelector("[data-uwe-stat=edges]").textContent = String(edges.length);
    workspace.querySelector("[data-uwe-stat=packages]").textContent = String(new Set(nodes.map(function (node) { return node.packageName; })).size);
    if (badge) badge.textContent = "Cytoscape + dagre active: wheel zoom, drag pan, click inspect";
    if (nodes[0]) setInspector(workspace, nodes[0]);
    if (nodes[0]) setActiveNodeButton(workspace, nodes[0].id);
    if (nodes[0]) setActivePackageButton(workspace, nodes[0].packageName);
    cy.on("tap", "node", function (event) {
      var data = event.target.data();
      setInspector(workspace, data);
      setActiveNodeButton(workspace, data.id);
      setActivePackageButton(workspace, data.packageName);
    });
    workspace.querySelectorAll("[data-uwe-action]").forEach(function (button) {
      button.addEventListener("click", function () {
        var action = button.dataset.uweAction;
        if (action === "fit") cy.fit(undefined, 36);
        if (action === "layout") cy.layout({ name: window.cytoscapeDagre ? "dagre" : "grid", rankDir: "LR", nodeSep: 70, edgeSep: 24, rankSep: 132, fit: true, padding: 36 }).run();
        if (action === "workspace-focus") setFocusMode(workspace, !workspace.classList.contains("uwe-focus-mode"), cy);
        if (action && action.indexOf("package:") === 0) {
          var packageName = action.slice("package:".length);
          var collection = cy.nodes().filter(function (node) { return node.data("packageName") === packageName; });
          if (collection.length > 0) cy.fit(collection.union(collection.connectedEdges()), 56);
          setActivePackageButton(workspace, packageName);
        }
      });
    });
    workspace.querySelectorAll("[data-uwe-focus-node]").forEach(function (button) {
      button.addEventListener("click", function () {
        var id = button.dataset.uweFocusNode;
        var node = nodes.find(function (candidate) { return candidate.id === id; });
        if (!node) return;
        setInspector(workspace, node);
        setActiveNodeButton(workspace, node.id);
        setActivePackageButton(workspace, node.packageName);
        var cyNode = cy.getElementById(id);
        if (cyNode && cyNode.length > 0) {
          cy.nodes().unselect();
          cyNode.select();
          cy.fit(cyNode.union(cyNode.connectedEdges()), 72);
        }
      });
    });
    var inspectorImage = workspace.querySelector("[data-uwe-inspector-image]");
    if (inspectorImage) inspectorImage.addEventListener("click", function () { openLightbox(workspace); });
    var openScreenshotButton = workspace.querySelector("[data-uwe-open-screenshot]");
    if (openScreenshotButton) openScreenshotButton.addEventListener("click", function () { openLightbox(workspace); });
    workspace.querySelectorAll("[data-uwe-lightbox-close]").forEach(function (button) {
      button.addEventListener("click", function () { closeLightbox(workspace); });
    });
    var lightbox = workspace.querySelector("[data-uwe-lightbox]");
    if (lightbox) {
      lightbox.addEventListener("click", function (event) {
        if (event.target === lightbox) closeLightbox(workspace);
      });
    }
    document.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        closeLightbox(workspace);
        if (workspace.classList.contains("uwe-focus-mode")) setFocusMode(workspace, false, cy);
      }
    });
  }

  initPanZoom();
  document.querySelectorAll(".uwe-engine-workspace").forEach(initWorkspace);
  document.documentElement.classList.add("uwe-workspace-active");
})();
